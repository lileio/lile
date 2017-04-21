# Pub Sub Spec

Pub Sub in Lile is a seperate library that can be used to augment an existing service to add asynchronous notifications which other services can listen to and act upon.

Each service should receive a message "at least once", meaning that if 6 copies of the `account_service` and 2 copies of the `audit_log_service` are running concurrently then one of each of these services should act upon a message that's been published, i.e one of the `account_service` nodes and one of the `audit_log_service` nodes. This can be thought of as "at least once per service".

In practise this works similarly to [Rabbit MQ's Pub Sub](https://www.rabbitmq.com/tutorials/tutorial-three-go.html) and [Google's Pub Sub](https://cloud.google.com/go/getting-started/using-pub-sub). Effectively each logical subscriber group has it's own queue which can pull messages and messages are distributed to all queues available.

Messages are in encoded in protobuf.

Tracing should be implemented by all implementations.
