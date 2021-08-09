package storage

import (
	"context"
	"fmt"
	"log"

	m "github.com/t-bfame/diago/pkg/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetTestLogs(ctx context.Context, testID m.TestID, testInstanceID m.TestInstanceID) ([]m.TestInstanceLog, error) {
	filter := bson.M{
		"testinstance_id": testInstanceID,
	}

	var testInstances []struct {
		Response string `bson:"response"`
	}

	projection := bson.D{
		{"response", 1},
	}

	cur, err := mongoClient.Database("diago-worker").Collection("responsedata").Find(ctx, filter, options.Find().SetProjection(projection))
	if err != nil {
		return nil, err
	}

	cur.All(ctx, &testInstances)
	log.Println(fmt.Sprintf("found test instances are : %s,  %b", testInstances, cur.Next(ctx)))
	var testInstanceLogs []m.TestInstanceLog
	for _, testInstance := range testInstances {
		testInstanceLogs = append(testInstanceLogs, m.TestInstanceLog(testInstance.Response))
	}
	return testInstanceLogs, nil
}
