package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"cloud.google.com/go/pubsub/v2"
	pb "cloud.google.com/go/pubsub/v2/apiv1/pubsubpb"
	"github.com/Desarrollo-de-Soluciones-Cloud/202612-MISW4204-Grupo11/internal/application/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PubSubClient struct {
	client     *pubsub.Client
	publisher  *pubsub.Publisher
	subscriber *pubsub.Subscriber
}

func NewPubSubClient(ctx context.Context, projectID, topicID, subscriptionID string) (*PubSubClient, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub new client: %w", err)
	}

	topicName := fmt.Sprintf("projects/%s/topics/%s", projectID, topicID)
	subName := fmt.Sprintf("projects/%s/subscriptions/%s", projectID, subscriptionID)

	if err := ensureTopicExists(ctx, client, topicName); err != nil {
		_ = client.Close()
		return nil, err
	}

	if err := ensureSubscriptionExists(ctx, client, topicName, subName); err != nil {
		_ = client.Close()
		return nil, err
	}

	return &PubSubClient{
		client:     client,
		publisher:  client.Publisher(topicID),
		subscriber: client.Subscriber(subscriptionID),
	}, nil
}

func ensureTopicExists(ctx context.Context, client *pubsub.Client, topicName string) error {
	_, err := client.TopicAdminClient.GetTopic(ctx, &pb.GetTopicRequest{Topic: topicName})
	if err == nil {
		return nil
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("pubsub get topic: %w", err)
	}
	_, err = client.TopicAdminClient.CreateTopic(ctx, &pb.Topic{Name: topicName})
	if err != nil && status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("pubsub create topic: %w", err)
	}
	return nil
}

func ensureSubscriptionExists(ctx context.Context, client *pubsub.Client, topicName, subName string) error {
	_, err := client.SubscriptionAdminClient.GetSubscription(ctx, &pb.GetSubscriptionRequest{Subscription: subName})
	if err == nil {
		return nil
	}
	if status.Code(err) != codes.NotFound {
		return fmt.Errorf("pubsub get subscription: %w", err)
	}
	_, err = client.SubscriptionAdminClient.CreateSubscription(ctx, &pb.Subscription{
		Name:  subName,
		Topic: topicName,
	})
	if err != nil && status.Code(err) != codes.AlreadyExists {
		return fmt.Errorf("pubsub create subscription: %w", err)
	}
	return nil
}

func (p *PubSubClient) PublishWeeklyReportJob(ctx context.Context, job ports.WeeklyReportJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal weekly report job: %w", err)
	}

	result := p.publisher.Publish(ctx, &pubsub.Message{
		Data: body,
	})
	if _, err := result.Get(ctx); err != nil {
		return fmt.Errorf("publish weekly report job: %w", err)
	}

	return nil
}

func (p *PubSubClient) ConsumeWeeklyReportJobs(ctx context.Context, processor func(context.Context, ports.WeeklyReportJob) error) error {
	go func() {
		err := p.subscriber.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			var job ports.WeeklyReportJob
			if err := json.Unmarshal(msg.Data, &job); err != nil {
				log.Printf("invalid report job payload: %v", err)
				msg.Nack()
				return
			}

			if err := processor(ctx, job); err != nil {
				log.Printf("report job failed request_id=%s: %v", job.RequestID, err)
				msg.Nack()
				return
			}

			msg.Ack()
		})
		if err != nil && ctx.Err() == nil {
			log.Printf("pubsub receive error: %v", err)
		}
	}()

	return nil
}

func (p *PubSubClient) Close() error {
	p.publisher.Stop()
	return p.client.Close()
}
