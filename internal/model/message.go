package model

type Message struct {
	Type    int    `json:"type"`    // 1: 私聊, 2: 群聊, 3: 心跳
	Target  string `json:"target"`  // 接收者的 UserID
	From    string `json:"from"`    // 发送者的 UserID
	Content string `json:"content"` // 消息内容
}
