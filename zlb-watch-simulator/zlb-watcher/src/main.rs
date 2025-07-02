//use std::collections::HashMap;
use std::sync::Arc;
use tokio::net::UdpSocket;
use tokio::sync::Mutex;
use log::{info, error};
use std::time::Duration;

mod types;
mod state;

use types::IncomingMessage;
use state::StateManager;

#[tokio::main]
async fn main() -> std::io::Result<()> {
    env_logger::init();
    
    let state_manager = Arc::new(Mutex::new(StateManager::new()));
    let socket = Arc::new(UdpSocket::bind("0.0.0.0:10000").await?);
    
    info!("ZLB-Watch node listening on 0.0.0.0:10000");

    // Task for receiving messages
    let recv_socket = socket.clone();
    let recv_state = state_manager.clone();
    tokio::spawn(async move {
        let mut buf = [0; 1024];
        loop {
            match recv_socket.recv_from(&mut buf).await {
                Ok((len, _addr)) => {
                    let msg_str = std::str::from_utf8(&buf[..len]).unwrap_or("");
                    match serde_json::from_str::<IncomingMessage>(msg_str) {
                        Ok(msg) => {
                            let mut state = recv_state.lock().await;
                            match msg {
                                IncomingMessage::SourceTx { tx_hash } => state.handle_source_tx(tx_hash),
                                IncomingMessage::SourceBlock { tx_hashes } => state.handle_source_block(tx_hashes),
                                IncomingMessage::AuditTx { source_tx_hash, committer } => state.handle_audit_tx(source_tx_hash, committer),
                                IncomingMessage::DestMint { source_tx_hash } => state.handle_dest_mint(source_tx_hash),
                            }
                        }
                        Err(e) => error!("Failed to parse message: {}", e),
                    }
                }
                Err(e) => error!("Error receiving UDP message: {}", e),
            }
        }
    });

    // Task for checking timeouts
    let check_state = state_manager.clone();
    tokio::spawn(async move {
        let mut interval = tokio::time::interval(Duration::from_secs(5));
        loop {
            interval.tick().await;
            let mut state = check_state.lock().await;
            let failures = state.check_for_failures();
            for (tx_hash, reason) in failures {
                match reason {
                    types::FailureReason::Omission => {
                        println!("\n>>> Proof of Bridge Omission generated for TX: {} <<<\n", tx_hash);
                    }
                    types::FailureReason::Finalization(committer) => {
                        println!("\n>>> Proof of Destination Omission generated for TX: {} (Committer: {}) <<<\n", tx_hash, committer);
                    }
                }
            }
        }
    });

    // Keep the main task alive
    tokio::signal::ctrl_c().await?;
    info!("Shutting down ZLB-Watch node.");
    Ok(())
}