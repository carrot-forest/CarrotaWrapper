package db

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type dbHistory struct {
	Agent     string `bson:"agent"`
	Group_id  string `bson:"group_id"`
	User_id   string `bson:"user_id"`
	User_name string `bson:"user_name"`
	Question  string `bson:"question"`
	Answer    string `bson:"answer"`
	Time      int64  `bson:"time"`
}

type WrapperRequest struct {
	Agent             string   `json:"agent"`
	Group_id          string   `json:"group_id"`
	Group_name        string   `json:"group_name"`
	User_id           string   `json:"user_id"`
	User_name         string   `json:"user_name"`
	Message           string   `json:"message"`
	Time              int64    `json:"time"`
	Original_response []string `json:"original_response"`
}

var Mongo = InitMongo()

func InitMongo() *mongo.Collection {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().SetAuth(options.Credential{
		Username: "root",
		Password: "root",
	}).ApplyURI("mongodb://127.0.0.1:27017"))
	if err != nil {
		log.Println("Connection MongoDB Error:", err)
		return nil
	}
	return client.Database("wrapper").Collection("history")
}

// func GetAll() ([]string, error) {
// 	var history []dbHistory
// 	var content []string
// 	cur, err := coll.Find(context.TODO(), nil)
// 	if err != nil {
// 		fmt.Println("db find err:", err)
// 		return nil, err
// 	}
// 	cur.All(context.TODO(), history)
// 	for _, oneHistory := range history {
// 		content = append(content, oneHistory.user+"说"+oneHistory.question)
// 		content = append(content, oneHistory.answer)
// 	}
// 	fmt.Println(content)
// 	return content, nil
// }

func GetHistory(req *WrapperRequest) ([]string, error) {
	var content []string

	// 随机获取一条数据
	randTime := req.Time - 3600 - rand.Int63n(7200)

	oneFilter := bson.M{
		"agent":    req.Agent,
		"group_id": req.Group_id,
		"time":     bson.M{"$gte": randTime, "$lt": req.Time - 3600}, // 只找群聊中最近 1 个小时的消息
	}

	// 找群聊中最近的 10 条数据
	filter := bson.M{
		"agent":    req.Agent,
		"group_id": req.Group_id,
		"time":     bson.M{"$gte": req.Time - 3600}, // 只找群聊中最近 1 个小时的消息
	}
	if req.Group_id == "" {
		oneFilter = bson.M{
			"agent":    req.Agent,
			"group_id": "",
			"user_id":  req.User_id,
			"time":     bson.M{"$gte": randTime, "$lt": req.Time - 3600}, // 只找群聊中最近 1 个小时的消息
		}
		filter = bson.M{
			"agent":    req.Agent,
			"group_id": "",
			"user_id":  req.User_id,
			"time":     bson.M{"$gte": req.Time - 3600}, // 只找群聊中最近 1 个小时的消息
		}
	}

	oneHistory := new(dbHistory)
	err := Mongo.FindOne(context.Background(), oneFilter).Decode(oneHistory)
	if err == nil {
		content = append(content, oneHistory.User_name+"说:"+oneHistory.Question)
		content = append(content, oneHistory.Answer)
	}

	findOptions := options.Find()
	findOptions.SetLimit(10)
	findOptions.SetSort(bson.D{{Key: "time", Value: -1}})

	cur, err := Mongo.Find(context.Background(), filter, findOptions) // 获取游标 cur
	if err != nil {
		fmt.Println("db find err:", err)
		return nil, err
	}
	for cur.Next(context.Background()) { // 遍历 cur 获得所有记录，生成 content 为一条 question 一条 answer
		oneHistory := new(dbHistory)
		cur.Decode(oneHistory)
		content = append(content, oneHistory.User_name+"说:"+oneHistory.Question)
		content = append(content, oneHistory.Answer)
	}

	return content, nil
}

func InsertHisory(req *WrapperRequest, answer string) {
	oneHistory := dbHistory{
		Agent:     req.Agent,
		Group_id:  req.Group_id,
		User_id:   req.User_id,
		User_name: req.User_name,
		Question:  req.Message,
		Answer:    answer,
		Time:      req.Time,
	}

	fmt.Printf("插入数据:\n %+v\n", oneHistory)
	_, err := Mongo.InsertOne(context.Background(), oneHistory)
	if err != nil {
		fmt.Println("insert err:", err)
		return
	}
}
