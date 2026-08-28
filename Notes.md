### Understanding
So if my understanding is correct, we have a production scheduling SaaS, uh, which schedules a different kind of, uh, productions and does all the different kinds of, uh, details or processing for the production. Um, this data needs to be sent to the ERP or the dashboard, which is basically on the client side or the customer side. Now the problem is when sub scheduling event happens, uh, what are the ways that we can, uh, send this data? Or how how can the the two systems communicate. There are two systems. There are two architectures or two philosophies, which is one is one is pull, which is, like, a HTTP request where the client requests and pulls the required response from the server. This is what a traditional approach would be. The other one is push where the client does not request. The client just keeps, you know, doing whatever it is doing. The server, when the event happens on the server side, it will automatically push the relevant response to the subscribed client. Is this correct? is... and the webhook is basically the second thing that I explained. It's just a program which waits for certain events to happen. And when certain events happen, it will push to the customers who are waiting for those events. Is this is this right?

### Refinements
- Pull (polling): client repeatedly asks the server "anything new?" Correct as you said.

- Push (webhooks): correct in spirit, but one detail — it's not literally the server reaching into the client's memory. The server makes a normal outbound HTTP POST to a URL the client gave it in advance. The client's side just needs a small HTTP server listening on that URL to receive it. So "push" here just means who initiates the HTTP request is inverted — server-initiated instead of client-initiated. Both directions are still plain HTTP.

- So the full chain: scheduling engine does its work → emits an internal event → webhook service picks it up → looks up registered endpoints for that customer/event type → delivers via HTTP POST to each. Issues #1-#5 in this project are entirely about that last stage — the delivery mechanics (queuing, worker pool, ordering, shutdown, failure visibility) — not the scheduling engine itself, which is assumed to already exist and just hand off events. 
    - So if the server initiates that CTP post to a predefined URL given by the client, what does... how does the client get the data? like, the HTTP server keeps listening. And when a post happens, what will happen? We'll give the data to client.
        - The client (Customer B) runs their own small HTTP server, always running, listening on that URL (e.g. https://customer-b.com/hooks/schedule-events). It sits idle, doing nothing, until a request arrives.

        - When our webhook service sends the POST, it's a normal HTTP request — headers, and a body containing the event data (JSON, typically something like {"event": "work_order.status_changed", "work_order_id": 123, "new_status": "delayed"}). That request travels over the internet and hits the client's listening server.


## Important - v1 does not external queues for storing the events it will only use golang.

- But v2 in the future will definitely use kafka.


## Issue #1 design — full record

Full decision log (initial proposal → issue found → final call) for every open
question on issue #1: `docs/adr/0001-accept-and-route-event.md`.

Architecture sketch (v1, with the Redis path marked as v2):

![architecture sketch](docs/adr/assets/0001-architecture-sketch.png)

## Article
- https://dev.to/vikthurrdev/designing-a-webhook-service-a-practical-guide-to-event-driven-architecture-3lep

- https://medium.com/@shivanimutke2501/day-50-system-design-concept-web-hooks-14615bd717a3