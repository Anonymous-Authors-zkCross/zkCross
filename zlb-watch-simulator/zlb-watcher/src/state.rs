use std::collections::HashMap;
use std::time::{Duration, Instant};
use crate::types::{TransactionState, FailureReason};
use log::{info, warn};

const OMISSION_TIMEOUT: Duration = Duration::from_secs(20);
const FINALIZATION_TIMEOUT: Duration = Duration::from_secs(30);

pub struct StateManager {
    watched_txs: HashMap<String, TransactionState>,
}

impl StateManager {
    pub fn new() -> Self {
        Self {
            watched_txs: HashMap::new(),
        }
    }

    pub fn handle_source_tx(&mut self, tx_hash: String) {
        info!(" New TX received, pending source confirmation: {}...", &tx_hash[..8]);
        self.watched_txs.entry(tx_hash).or_insert_with(TransactionState::new);
    }

    pub fn handle_source_block(&mut self, tx_hashes: Vec<String>) {
        for tx_hash in tx_hashes {
            if let Some(state) = self.watched_txs.get_mut(&tx_hash) {
                if let TransactionState::PendingSourceConfirmation {..} = state {
                    info!(" TX confirmed on source, pending audit: {}...", &tx_hash[..8]);
                    *state = TransactionState::PendingAudit {
                        deadline: Instant::now() + OMISSION_TIMEOUT,
                    };
                }
            }
        }
    }

    pub fn handle_audit_tx(&mut self, tx_hash: String, committer: String) {
        if let Some(state) = self.watched_txs.get_mut(&tx_hash) {
            if let TransactionState::PendingAudit {..} = state {
                info!(" TX audited by {}, pending destination confirmation: {}...", committer, &tx_hash[..8]);
                *state = TransactionState::PendingDestinationConfirmation {
                    committer,
                    deadline: Instant::now() + FINALIZATION_TIMEOUT,
                };
            }
        }
    }

    pub fn handle_dest_mint(&mut self, tx_hash: String) {
        if self.watched_txs.remove(&tx_hash).is_some() {
            info!(" TX finalized on destination, removing from watch: {}...", &tx_hash[..8]);
        }
    }

    pub fn check_for_failures(&mut self) -> Vec<(String, FailureReason)> {
        let mut failures = Vec::new();
        let now = Instant::now();

        self.watched_txs.retain(|tx_hash, state| {
            match state {
                TransactionState::PendingAudit { deadline } => {
                    if now > *deadline {
                        warn!(" Omission timeout for TX: {}...", &tx_hash[..8]);
                        failures.push((tx_hash.clone(), FailureReason::Omission));
                        return false; // Remove from map
                    }
                }
                TransactionState::PendingDestinationConfirmation { committer, deadline } => {
                    if now > *deadline {
                        warn!(" Finalization timeout for TX: {}... (Committer: {})", &tx_hash[..8], committer);
                        failures.push((tx_hash.clone(), FailureReason::Finalization(committer.clone())));
                        return false; // Remove from map
                    }
                }
                _ => {}
            }
            true // Keep in map
        });

        failures
    }
}