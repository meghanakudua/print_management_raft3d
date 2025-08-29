# Distributed Print Management System - Raft3D
**Tech Stack:** Go, Raft Consensus Algorithm, REST API, BoltDB
- Implemented a fault-tolerant distributed system using the Raft consensus protocol
- Created REST APIs to handle print jobs and manage cluster
- Allowed new nodes to join the system and used BoltDB to save data even after restarts
- Simulated a shared print queue that continues working even if some nodes fail
