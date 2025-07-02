import socket
import json
import random

# Network configuration
HOST, PORT = "0.0.0.0", 10002
ZLB_HOST, ZLB_PORT = "zlb-watcher", 10000

sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
sock.bind((HOST, PORT))

print("Destination Chain Simulator Started")

def send_message(host, port, message):
    sock.sendto(json.dumps(message).encode('utf-8'), (host, port))

while True:
    data, addr = sock.recvfrom(1024)
    audit_tx = json.loads(data.decode('utf-8'))
    
    if audit_tx.get("type") == "AuditTx":
        tx_hash = audit_tx["source_tx_hash"]
        print(f"Received Audit for TX {tx_hash[:8]}...")
        
        # Simulate potential finalization failure
        if random.random() < 0.8: # 80% chance to finalize
            mint_event = {
                "type": "DestMint",
                "source_tx_hash": tx_hash
            }
            send_message(ZLB_HOST, ZLB_PORT, mint_event)
            print(f"  -> Minted TX {tx_hash[:8]}...")
        else:
            print(f"  -> FAILED TO MINT TX {tx_hash[:8]}...")