use serde::{Deserialize, Serialize};
use std::time::{Instant};


#[derive(Serialize, Deserialize, Debug)]
#[serde(tag = "type")]
pub enum IncomingMessage {
    SourceTx { tx_hash: String },
    SourceBlock { tx_hashes: Vec<String> },
    AuditTx { source_tx_hash: String, committer: String },
    DestMint { source_tx_hash: String },
}


pub enum TransactionState {
    PendingSourceConfirmation { received_at: Instant },
    PendingAudit { deadline: Instant },
    PendingDestinationConfirmation { committer: String, deadline: Instant },
}

impl TransactionState {
    pub fn new() -> Self {
        TransactionState::PendingSourceConfirmation {
            received_at: Instant::now(),
        }
    }
}


pub enum FailureReason {
    Omission,
    Finalization(String),
}