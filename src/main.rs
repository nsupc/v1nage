use clap::Parser;
use env_logger::Env;
use eventsource_client as es;
use eventsource_client::Client;
use futures::TryStreamExt;
use regex::Regex;
use reqwest::blocking::{Body, ClientBuilder};
use serde::{Deserialize, Serialize};
use serde_json;
use std::time::Duration;

#[derive(Debug, Parser)]
struct Cli {
    #[clap(short, long, help = "NationStates API user agent")]
    user_agent: String,
    #[clap(short, long)]
    region: String,
    #[clap(short, long)]
    webhook_url: String,
    #[clap(short, long, help = "Telegram template to autofill")]
    telegram_template: Option<String>,
}

#[derive(Debug, Deserialize)]
struct Event {
    #[serde(rename = "str")]
    text: String,
}

#[derive(Serialize)]
struct Response {
    content: String,
}

impl Into<Body> for Response {
    fn into(self) -> Body {
        Body::from(self.content)
    }
}

#[tokio::main(flavor = "current_thread")]
async fn main() -> Result<(), anyhow::Error> {
    env_logger::init_from_env(Env::default().default_filter_or("info"));

    let args = Cli::parse();

    let region = args.region.to_lowercase().replace(" ", "_");
    let sse_url = format!("https://nationstates.net/api/region:{}", &region);

    let wa_regex = Regex::new(r"^@@(.*)@@ was admitted to the World Assembly\.$")?;
    let update_regex = Regex::new(&format!("^%%{}%% updated.$", &region))?;

    // let webhook_client = WebhookClient::new(&args.webhook_url);
    let http_client = ClientBuilder::new().build()?;

    let client = es::ClientBuilder::for_url(&sse_url)?
        .header("User-Agent", &args.user_agent)?
        .reconnect(
            es::ReconnectOptions::reconnect(true)
                .retry_initial(true)
                .delay_max(Duration::from_secs(120))
                .build(),
        )
        .build();

    let mut stream = client
        .stream()
        .map_ok(|event| match event {
            es::SSE::Connected(_) => {
                log::info!("connected")
            }
            es::SSE::Event(ev) => {
                let event: Event = serde_json::from_str(&ev.data).unwrap();

                log::info!("event: {:?}", event);

                if update_regex.is_match(&event.text) {
                    let resp = Response {
                        content: format!("{} updated", &region),
                    };

                    match http_client.post(&args.webhook_url).body(resp).send() {
                        Ok(_resp) => {
                            log::info!("posted");
                        }
                        Err(e) => {
                            log::error!("error sending webhook: {}", e);
                        }
                    }
                    // futures::executor::block_on(async {
                    //     match webhook_client
                    //         .send(|message| message.content(&format!("{} updated!", &region))).await {
                    //         Ok(_) => log::info!("updated"),
                    //         Err(err) => log::error!("{}", err),
                    //     }
                    // })
                };

                if let Some(captures) = wa_regex.captures(&event.text) {
                    let nation_name = captures.get(0).unwrap().as_str().to_string();

                    let content = if args.telegram_template.is_some() {
                        format!("New WA nation: {} ([endorse](https://www.nationstates.net/nation={}#composebutton), [telegram](https://www.nationstates.net/page=compose_telegram?tgto={}&message=%25{}%25&generated_by=waatcher))", &nation_name, &nation_name, &nation_name, args.telegram_template.as_ref().unwrap().replace("%", ""))
                    } else {
                        format!("New WA nation: {} ([endorse](https://www.nationstates.net/nation={}#composebutton))", &nation_name, &nation_name)
                    };

                    let resp = Response {
                        content,
                    };

                    match http_client.post(&args.webhook_url).body(resp).send() {
                        Ok(_resp) => {
                            log::info!("posted");
                        }
                        Err(e) => {
                            log::error!("error sending webhook: {}", e);
                        }
                    }

                    // futures::executor::block_on(async {
                    //     match webhook_client.send(|message| {
                    //         let content = if args.telegram_template.is_some() {
                    //             format!("New WA nation: {} ([endorse](https://www.nationstates.net/nation={}#composebutton), [telegram](https://www.nationstates.net/page=compose_telegram?tgto={}&message=%25{}%25&generated_by=waatcher))", &nation_name, &nation_name, &nation_name, args.telegram_template.as_ref().unwrap().replace("%", ""))
                    //         } else {
                    //             format!("New WA nation: {} ([endorse](https://www.nationstates.net/nation={}#composebutton))", &nation_name, &nation_name)
                    //         };
                    //
                    //         message.content(&content)
                    //     }).await {
                    //         Ok(_) => log::info!("updated"),
                    //         Err(err) => log::error!("{}", err),
                    //     }
                    // });
                }
            }
            _ => {}
        })
        .map_err(|err| log::error!("error streaming events: {:?}", err));

    while let Ok(Some(_)) = stream.try_next().await {}

    Ok(())
}
