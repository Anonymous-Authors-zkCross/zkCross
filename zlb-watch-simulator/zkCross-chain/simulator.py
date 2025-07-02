import socket
import json
import random

# Network configuration
HOST, PORT = "0.0.0.0", 10001
DEST_HOST, DEST_PORT = "dest-chain", 10002
ZLB_HOST, ZLB_PORT = "zlb-watcher", 10000

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock.bind((HOST, PORT))

print("zkCross Simulator Started")

def send_message(host, port, message):
    sock.sendto(json.dumps(message).encode('utf-8'), (host, port))

while True:
    data, addr = sock.recvfrom(1024)
    block = json.loads(data.decode('utf-8'))
    
    if block.get("type") == "SourceBlock":
        print(f"Received Source Block {block['block_height']}")
        for tx_hash in block["tx_hashes"]:
            # Simulate potential censorship
            if random.random() < 0.9: # 90% chance to process honestly
                audit_tx = {
                    "type": "AuditTx",
                    "source_tx_hash": tx_hash,
                    "committer": "0xCommitter"
                }
                send_message(DEST_HOST, DEST_PORT, audit_tx)
                send_message(ZLB_HOST, ZLB_PORT, audit_tx)
                print(f"  -> Audited TX {tx_hash[:8]}...")
            else:
                print(f"  -> OMITTED TX {tx_hash[:8]}...")