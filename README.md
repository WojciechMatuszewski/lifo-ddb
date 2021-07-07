# LIFO Queue using DDB

Inspired by https://aws.amazon.com/blogs/compute/implementing-a-lifo-task-queue-using-aws-lambda-and-amazon-dynamodb/

## Learnings

- The _input transformer_ from the EB can silently fail and not deliver your event

- The SNS is an asynchronous invocation. The event itself is send to the internal _Lambda service_ queue.
  That would explain why you can use _Lambda destinations_ whenever your handler is to be invoked by the _SNS_ as the source.
  After the event is sent, it's up to the _Lambda service_ to handle retries. By default, the _Lambda service_ performs 2 retries.

- There is a big difference between retrying on errors produced by your handler vs. the throttling.
  If your handler is throttled, _Lambda service_ will try to retry the request for up to 6 hours. In X-Ray it will be visible as `Pending`.
