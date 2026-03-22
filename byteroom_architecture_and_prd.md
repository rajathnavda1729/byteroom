# **ByteRoom: Real-Time Tech & System Design Chat Platform**

**Document Type:** System Architecture & Product Requirements Document

**Target Audience:** Engineering Team / AI Coding Assistants (Cursor, Copilot)

## **1\. Project Overview**

**ByteRoom** is a real-time chat application specifically tailored for Computer Science, DSA (Data Structures & Algorithms), and System Design discussions. The name reflects its core purpose: a dedicated, low-latency "room" for engineers to discuss "bytes"—sharing complex code, evaluating algorithmic time complexities, and drawing distributed architecture diagrams without the formatting limitations of generic messaging apps.

The platform features strict message delivery guarantees and specialized handling for large code snippets, markdown, and visual whiteboard states.

Development is split into two phases:

* **Phase 1 (MVP):** Target 100 Daily Active Users (DAU). Focus on core domain features and rapid iteration.  
* **Phase 2 (Scale):** Target 1,000,000 DAU. Focus on distributed systems, high throughput, and advanced search.

## **2\. Core Requirements**

### **Functional Requirements (FR)**

1. **Messaging:** 1:1 and Group chat support.  
2. **Delivery Guarantee:** Messages marked as "Sent/Delivered" (status \= true) *must* be durably persisted.  
3. **Group Management:** Dynamic addition and removal of members in chat rooms.  
4. **History:** Users can securely retrieve chronologically ordered past messages.  
5. **Domain-Specific:** \* Native support for Markdown and multi-language syntax-highlighted code blocks.  
   * Support for text-to-diagram rendering (Mermaid.js).  
   * Support for Media (images) and interactive system design diagrams (e.g., Excalidraw JSON state).

### **Non-Functional Requirements (NFR)**

1. **Latency:** End-to-end message delivery in \< 500ms.  
2. **Durability:** Zero message loss once acknowledged to the sender.  
3. **Availability:** Highly available system (favoring partition tolerance during network blips).  
4. **Security:** Strict server-side XSS sanitization for code blocks.

## **3\. Phase 1: MVP Architecture (100 DAU)**

### **3.1 Overview**

A monolithic architecture designed for speed of development while safely validating domain-specific features.

* **Compute:** Single Monolith written in **Go (Golang)**. Leveraging Go's net/http and gorilla/websocket (or nhooyr.io/websocket). Go's lightweight goroutines make handling concurrent WebSocket connections highly efficient out of the box and sets a strong foundation for extracting the Edge servers in Phase 2\.  
* **Transport:** WebSockets for real-time delivery; REST for historic/media operations.  
* **Database:** PostgreSQL (Relational tables for Users, Groups, and append-only Messages).  
* **Storage:** AWS S3 (or equivalent) for diagram/image binary storage.

### **3.2 System Flow Diagrams (Phase 1\)**

#### **3.2.1 Core Message Delivery Flow**

sequenceDiagram  
    participant Sender  
    participant Go\_Monolith  
    participant PostgreSQL  
    participant Recipient  
      
    Sender-\>\>Go\_Monolith: WS: Send Message (JSON)  
    Go\_Monolith-\>\>Go\_Monolith: Validate & Sanitize (XSS)  
    Go\_Monolith-\>\>PostgreSQL: INSERT into messages  
    PostgreSQL--\>\>Go\_Monolith: Commit Success  
    Go\_Monolith--\>\>Sender: WS: ACK (status: true)  
      
    alt Recipient Connected  
        Go\_Monolith-\>\>Recipient: WS: Push Message  
        Recipient--\>\>Go\_Monolith: WS: Delivery ACK  
    else Recipient Offline  
        Go\_Monolith-\>\>Go\_Monolith: Skip (Stored in DB for next login)  
    end

#### **3.2.2 Media / Diagram Upload Flow (Pre-signed URL)**

sequenceDiagram  
    participant Client  
    participant Go\_API  
    participant S3 Bucket  
      
    Client-\>\>Go\_API: POST /api/upload/request (file\_meta)  
    Go\_API--\>\>Client: 200 OK { presigned\_url }  
    Client-\>\>S3 Bucket: PUT file binary (using URL)  
    S3 Bucket--\>\>Client: 200 OK  
    Client-\>\>Go\_API: WS: Send Message { media\_url, type: image }

## **4\. Phase 2: Scale Architecture (1M DAU)**

### **4.1 Overview**

An event-driven microservices architecture decoupling connection management, ingestion, and storage to handle massive concurrent WebSockets and high write throughput.

* **Edge:** Fleet of stateless WebSocket Servers (**Go**). Extracted directly from the Phase 1 monolith.  
* **Message Broker:** Apache Kafka (Guarantees ingestion durability before DB write).  
* **State/Routing:** Redis (Maps user\_id \-\> websocket\_server\_ip).  
* **Primary DB:** ScyllaDB / Cassandra (Wide-column store for heavy message writes/reads).  
* **Search Engine:** Elasticsearch (For indexing and searching code snippets/history).  
* **CDN:** Cloudflare / AWS CloudFront (For media delivery).

### **4.2 System Flow Diagrams (Phase 2\)**

#### **4.2.1 Distributed Message Routing Flow**

sequenceDiagram  
    participant Sender  
    participant WS\_Server\_A (Edge)  
    participant Kafka  
    participant Chat\_Worker  
    participant Redis  
    participant ScyllaDB  
    participant WS\_Server\_B (Edge)  
    participant Recipient

    Sender-\>\>WS\_Server\_A (Edge): WS: Send Message  
    WS\_Server\_A (Edge)-\>\>Kafka: Publish to 'chat-inbound'  
    Kafka--\>\>WS\_Server\_A (Edge): Broker ACK  
    WS\_Server\_A (Edge)--\>\>Sender: WS: ACK (status: true)  
      
    par Async Processing  
        Kafka-\>\>Chat\_Worker: Consume Message  
        Chat\_Worker-\>\>Redis: GET connection(Recipient\_ID)  
        Redis--\>\>Chat\_Worker: Return WS\_Server\_B\_IP  
        Chat\_Worker-\>\>WS\_Server\_B (Edge): RPC/Forward Message  
        WS\_Server\_B (Edge)-\>\>Recipient: WS: Push Message  
    and Async Storage  
        Kafka-\>\>ScyllaDB: Batch INSERT Message  
    and Async Search Indexing  
        Kafka-\>\>Elasticsearch: Index Code/Text  
    end

## **5\. Frontend Architecture**

Because ByteRoom handles complex state (live WebSocket connections) and domain-specific rendering (Markdown, syntax highlighting, Mermaid, and interactive diagrams), the frontend requires a robust Single Page Application (SPA) architecture.

### **5.1 Tech Stack**

* **Framework:** React 18+ with TypeScript (via Vite for fast HMR and optimized builds).  
* **Styling:** Tailwind CSS (for rapid UI development and easy dark mode support, crucial for developer tools).  
* **State Management:** Zustand (for lightweight, boilerplate-free global state handling of Chat Rooms, Active Users, and WebSocket connection status).  
* **Local Caching:** IndexedDB (via idb or localforage) to cache historical messages and allow offline viewing of previous study sessions.

### **5.2 Domain-Specific Rendering Libraries**

To support the CS/DSA requirements, the frontend will integrate the following key libraries:

* **Markdown, Code & Mermaid:** \* react-markdown for safely parsing Markdown text.  
  * react-syntax-highlighter for syntax highlighting of standard code blocks.  
  * mermaid API to intercept code blocks tagged with mermaid and render them natively into SVG architecture diagrams within the chat bubble.  
* **System Design Diagrams:**  
  * @excalidraw/excalidraw (React component). When a message payload contains content\_type: "diagram\_state", the frontend renders a collaboratively editable Excalidraw canvas inline.

### **5.3 WebSocket Connection Management**

* **Connection Lifecycle:** A dedicated Zustand store or custom React Hook (useWebSocket) will manage the connection.  
* **Resilience:** Implement exponential backoff for auto-reconnection if the socket drops.  
* **Optimistic UI:** When a user hits "Send", the message immediately appears in the UI with a distinct styling (e.g., greyed out). Once the server sends the ACK, the UI updates to its "Delivered" state.

## **6\. Data Contracts & Payloads**

### **6.1 Standard Text, Code, & Mermaid Message Payload (WebSocket)**

*Notice how Mermaid fits seamlessly into the standard markdown payload as a code block.*

{  
  "event": "message.send",  
  "data": {  
    "message\_id": "msg\_01HQ...",  
    "chat\_id": "group\_101",  
    "sender\_id": "user\_404",  
    "content\_type": "markdown",  
    "content": "Here is the load balancer architecture:\\n\`\`\`mermaid\\ngraph TD;\\n    Client--\>LB;\\n    LB--\>Server1;\\n    LB--\>Server2;\\n\`\`\`",  
    "timestamp": "2026-03-21T14:40:23Z"  
  }  
}

### **6.2 Interactive Diagram Payload**

*For embedding editable whiteboard states (Excalidraw).*

{  
  "event": "message.send",  
  "data": {  
    "message\_id": "msg\_01HS...",  
    "chat\_id": "group\_101",  
    "sender\_id": "user\_404",  
    "content\_type": "diagram\_state",  
    "content": "{ \\"type\\": \\"excalidraw\\", \\"version\\": 2, \\"elements\\": \[...\] }",  
    "timestamp": "2026-03-21T14:50:00Z"  
  }  
}

## **7\. Implementation Notes for AI / Developers**

1. **Security First:** The frontend should NOT be fully trusted to sanitize HTML. The backend must sanitize Markdown content (e.g., using Go's bluemonday) to strip \<script\> and \<iframe\> tags before broadcasting to prevent XSS attacks via code blocks. Ensure the sanitization policy allows standard Markdown elements but strips dangerous HTML attributes.  
2. **Mermaid Rendering:** Ensure the frontend dynamically imports the mermaid package only when a Mermaid code block is detected to keep the initial client bundle size small.  
3. **Idempotency:** The backend MUST check message\_id against the database/cache. If the client retries sending a message due to a network timeout, the backend should return an ACK but *not* duplicate the database entry.  
4. **Heartbeat (Ping/Pong):** Ensure the frontend periodically sends a lightweight ping, or responds to backend pings, to keep the WebSocket connection alive through load balancers and prevent false "offline" indicators.