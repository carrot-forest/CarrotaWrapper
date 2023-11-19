# Wrapper

## 目录结构
    .
    ├── handle                  
    │   ├── chat                - 调用 qwen 服务
    │   │   ├── chat.go         
    │   │   └── chat_url.txt    - qwen 服务的 url
    │   ├── db
    │   │   └── db.go           - MongoDB 数据库
    │   ├── go.mod              
    │   ├── go.sum              
    │   └── main.go             - wrapper 服务
    └── prompt.txt              - wrapper 的 prompt
## [qwen.cpp](https://github.com/QwenLM/qwen.cpp)
Wrapper 使用的大语言模型为阿里的开源大模型 Qwen，由于我笔记本的性能问题，需要使用官方提供的 qwen.cpp 版本。

qwen.cpp 是官方将模型转化为 ggml 库可以支持的格式，可以压缩模型的大小。ggml 库对 CPU 的矩阵运算进行了优化，同时也支持 GPU 的加速。使用 qwen.cpp 可以在显存比较小的情况下运行模型。

### 环境
python 3.8及以上版本，我选用了Python 3.8

CUDA 11.4及以上，我选用了 cuda 11.8

pytorch 1.12及以上版本，推荐 2.0 及以上版本，我根据 pytorch 官网选择 cuda118 直接安装对应版本。

### 模型部署
    git clone --recursive https://github.com/QwenLM/qwen.cpp && cd qwen.cpp

然后在 qwen 提供的模型库中下载模型，包括完整模型文件和 qwen.tiktoken 文件。

接着通过仓库中的 convert.py 将完整模型转化为 ggml 模型，这里下载的是 Qwen-14B-chat 的完整模型，然后转化为 int4 量化的 ggml 模型，根据仓库的参数说明，进行下面命令。

    python3 qwen_cpp/convert.py -i Qwen/Qwen-14B-Chat -t q4_0 -o qwen14b-q4-ggml.bin

之后是编译和运行，使用 cmake 工具编译。

    cmake -B build -DGGML_CUBLAS=ON
    //cmake 参数加入 cublas，在预编译后提示中要能找到 cublas。
    cmake --build build -j
    //编译成可执行文件
    ./build/bin/main -m qwen14b-q4-ggml.bin --tiktoken qwen.tiktoken -p 你好
    //运行模型，由于编译时加入了 cublas 参数，会使用 GPU 加速。

### 修改qwen.cpp文件
因为 main.cpp 实际上至少对使用 qwen.cpp 做了封装，可以对qwen.cpp 进行一些简单的修改，例如加入设置 prompt 的环节。
在代码的 ``build_prompt`` 函数中将外部 prompt.txt 文件读入即可。

### python binding
因为 c++ 编写 web 服务有点麻烦，而官方提供了 python binding 的方法，可以在 Python 中调用 qwen.cpp 库，而 Python 编写 web 服务比较简单。这里我编写了一个 API.py, 它在本地 48283 端口提供一个服务，接收下面格式的 json 。
```json
{
    "history": ["你好","我是卡洛，你好！"],
    "prompt": "prompt_wrapper.txt",
    "query": "今天天气很好呢",
}
```
会返回一个字符串 response 表示回复内容。
### prompt 设计
Wrapper 其实要实现两个功能:
1.  当上层提供了原始回复时，需要进行包装后输出。
2.  上层没有提供原始回复，根据用户的提问直接回复。

总体设计基于 hjy 友情提供的 prompt，主要是给 qwen 一个卡洛的身份，并告知它的性格等特征。
同时告知它具有一个任务，当识别到 "wrapper:" 后，将后面的内容进行改写，这样实现了功能1。


## API

### [POST]/wrapper

#### Request
```
{
    "agent" : "feishu",
    "group_id" : "926170830",
    "group_name" : "软工交流群",
    "user_id" : "13421892120",
    "user_name" : "Bunny",   
    "message" : "你好",
    "time" : 1699806329,
    "original_response" : "你好鸭！",
}
```
1. agent: 客户端，目前可选为"feishu" (string) 
2. group_id: 飞书群唯一标识 (strinig)
3. group_name: 群聊名称 (string)
4. user_id: 用户的飞书唯一标识 (string)，若该字段为空串，则没有用户发消息，只需要对卡洛的原始回复进行包装（定时回复时可能需要用到）此时忽略 user_name, message 字段。
5. user_name: 用户的飞书昵称 (string)
6. message: 用户发的消息 (string)
7. time: 用户发送的消息的时间戳(秒为单位) (int64)
8. original_response: 卡洛的原始回复 (string)，若该字段为空串，则卡洛没有原始回复，直接根据用户消息生成回复。

#### Response
```json
{
    "response" : "我是卡洛，你好！"
}
```
1. *response:* 卡洛最终的回复。

## MongoDB
### 安装
https://www.mongodb.com/docs/manual/tutorial/install-mongodb-on-ubuntu/
### 启动 MongoDB
    sudo systemctl start mongod 
    启动
    sudo systemctl status mongod
    查看状态
    sudo systemctl stop mongod
    关闭
    sudo systemctl restart mongod
    重启
    mongosh
    数据库命令行
### 用户
现在有两个用户 

    admin>show users
    [
    {
        _id: 'admin.root',
        userId: new UUID("0dc76842-4a15-4363-8dac-8d63a9ea9029"),
        user: 'root',
        db: 'admin',
        roles: [ { role: 'root', db: 'admin' } ],
        mechanisms: [ 'SCRAM-SHA-1', 'SCRAM-SHA-256' ]
    },
    {
        _id: 'admin.test',
        userId: new UUID("9d3d06e1-51ec-411f-911f-31f6a7e0f7d5"),
        user: 'test',
        db: 'admin',
        roles: [ { role: 'root', db: 'admin' } ],
        mechanisms: [ 'SCRAM-SHA-1', 'SCRAM-SHA-256' ]
    }
    ]
密码与用户名相同。
### MongoDB web gui
使用 [MongoDB PHP GUI](https://github.com/SamuelTallet/MongoDB-PHP-GUI).
#### 环境配置
直接用 docker

    docker pull samueltallet/mongodb-php-gui
#### 运行
不作修改

    docker run --add-host localhost:172.17.0.1 --publish 5000:5000 --rm samueltallet/mongodb-php-gui
这个 gui 好像我会出现一个找不到数据库服务的问题，暂时不清楚问题原因。
#### 查看
打开 http://127.0.0.1:5000/ 查看

