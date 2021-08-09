package storage

import (
	"context"

	m "github.com/t-bfame/diago/pkg/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetTestLogs(ctx context.Context, testID m.TestID, testInstanceID m.TestInstanceID) (map[m.JobID][]m.TestInstanceLog, error) {
	filter := bson.M{
		"testinstance_id": testInstanceID,
	}

	var testInstances []struct {
		Response string `bson:"response"`
		JobID    string `bson:"job_id"`
	}

	projection := bson.D{
		{"response", 1},
		{"job_id", 2},
	}

	cur, err := mongoClient.Database("diago-worker").Collection("responsedata").Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}

	cur.All(ctx, &testInstances)

	var logsByID = make(map[m.JobID][]m.TestInstanceLog)
	for _, testInstance := range testInstances {
		jobID := m.JobID(testInstance.JobID)
		logsByID[jobID] = append(logsByID[jobID], m.TestInstanceLog(testInstance.Response))
	}

	return logsByID, nil
}
