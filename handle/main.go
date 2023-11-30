package main

import (
	"fmt"
	"handle/chat"
	"handle/db"
	"net/http"
	"strings"
	"os"
	"github.com/gin-gonic/gin"
)

func wrapper(c *gin.Context) {
	req := db.WrapperRequest{}
	c.BindJSON(&req)
	fmt.Printf("收到 request:\n %+v\n", req)

	var resp []string // 卡洛回复内容
	no_response := req.Original_response == nil || len(req.Original_response) == 0 || req.Original_response[0] == ""
	history, err := db.GetHistory(&req)
	if err != nil {
		fmt.Println("GetHistory err:", err)
		c.JSON(http.StatusOK, gin.H{
			"response": []string{"卡洛看起来失踪了喵，可能是跑出去玩了!"},
		})
		return
	}
	fmt.Println("历史消息: ", history)

	//db.GetAll()
	if no_response { // 没有原始回复，直接聊天
		if req.Message == "" {
			fmt.Println("no Message and need chat")
			c.JSON(http.StatusOK, gin.H{
				"response": []string{"卡洛不知道你在说什么喵"},
			})
			return
		}
		prompt_content, err := os.ReadFile("./prompt.txt")
	
		if err != nil {
			fmt.Println("can't not find prompt!",err);
			c.JSON(http.StatusOK, gin.H{
				"response": []string{"卡洛看起来失踪了喵，可能是跑出去玩了!"},
			})
			return
		}
		prompt := string(prompt_content)
		chatData := chat.ChatRequest{
			History: history,
			Prompt:  prompt,
			Query:   req.User_name + "说:" + req.Message,
		}
		content, err := chatData.Chat()
		if err != nil {
			fmt.Println("chat err:", err)
			c.JSON(http.StatusOK, gin.H{
				"response": []string{"卡洛看起来失踪了喵，可能是跑出去玩了!"},
			})
			return
		}
		fmt.Println(content)
		db.InsertHisory(&req, content)
		resp = append(resp, content)
	} else {
		
		fmt.Println("wrapper!!!!")
		prompt_content, err := os.ReadFile("./prompt_wrapper.txt")
	
		if err != nil {
			fmt.Println("can't not find prompt!",err);
			c.JSON(http.StatusOK, gin.H{
				"response": []string{"卡洛看起来失踪了喵，可能是跑出去玩了!"},
			})
			return
		}
		prompt := string(prompt_content)
		if strings.Contains(req.Original_response[0], "查询到的作业内容") {
			fmt.Println("作业！！")
			fmt.Printf(req.Original_response[0])
			chatData := chat.ChatRequest{
				History: nil,
				Prompt:  prompt,
				Query:   "wrapper: 快去加油写作业！",
			}
			content, err := chatData.Chat()
			if err != nil {
				fmt.Println("chat err:", err)
				c.JSON(http.StatusOK, gin.H{
					"response": []string{"卡洛看起来失踪了喵，可能是跑出去玩了!"},
				})
				return
			}
			req.Original_response[0] = req.Original_response[0] + content
			c.JSON(http.StatusOK, gin.H{
				"response": req.Original_response,
			})
			return
		}
		chatData := chat.ChatRequest{
			History: nil,
			Prompt:  prompt,
			Query:   "wrapper: " + req.Original_response[0],
		}
		content, err := chatData.Chat()
		if err != nil {
			fmt.Println("chat err:", err)
			c.JSON(http.StatusOK, gin.H{
				"response": []string{"卡洛看起来失踪了喵，可能是跑出去玩了!"},
			})
			return
		}
		db.InsertHisory(&req, content)
		resp = append(resp, content)
	}
	c.JSON(http.StatusOK, gin.H{
		"response": resp,
	})
}

func parser(c *gin.Context) {
	c.String(http.StatusOK, "{ \"plugin\": [{  \"id\": \"repeater\",  \"param\": null},{  \"id\": \"weather\",  \"param\": {    \"city\": \"北京\"  }},{  \"id\": \"divine\",  \"param\": null},{  \"id\": \"homework\",  \"param\": {    \"subject\": \"语文\",    \"isAddHomework\": \"false\",    \"Content\": \"作文，明天截止\",    \"Deadline\": \"明天\"  }}        ]      }")
}

func main() {
	r := gin.Default()
	r.POST("/wrapper", wrapper)
	r.POST("/parser", parser)
	r.Run("0.0.0.0:8283")
}
