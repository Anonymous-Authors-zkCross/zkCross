import socket
import time
import json
import random
import hashlib

# Network configuration
ZKCROSS_HOST, ZKCROSS_PORT = "zkcross-chain", 10001
ZLB_HOST, ZLB_PORT = "zlb-watcher", 10000

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)

def send_message(host, port, message):
    sock.sendto(json.dumps(message).encode('utf-8'), (host, port))

print("Source Chain Simulator Started")
tx_pool = []
block_height = 0

while True:
    # Generate a new transaction
    tx_hash = hashlib.sha256(str(random.random()).encode()).hexdigest()
    tx = {"type": "SourceTx", "tx_hash": tx_hash}
    tx_pool.append(tx_hash)
    
    # Send Attestation Request to ZLB-Watch
    send_message(ZLB_HOST, ZLB_PORT, tx)
    print(f"Sent TX {tx_hash[:8]}... to ZLB Watcher")
    
    # Every 5 transactions, create a block
    if len(tx_pool) >= 5:
        block = {
            "type": "SourceBlock",
            "block_height": block_height,
            "tx_hashes": tx_pool
        }
        # Send block to zkCross and ZLB-Watch
        send_message(ZKCROSS_HOST, ZKCROSS_PORT, block)
        send_message(ZLB_HOST, ZLB_PORT, block)
        print(f"--- Created Block {block_height} with {len(tx_pool)} TXs ---")
        
        tx_pool = []
        block_height += 1

    time.sleep(2) # Generate a tx every 2 seconds