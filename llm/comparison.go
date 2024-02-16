package llm

import (
	"fmt"
	"os"
	"script_validation/nomicai"
	"script_validation/pocketbase"
	"script_validation/utils"

	"github.com/joho/godotenv"
	"github.com/sashabaranov/go-openai"
)

var messages = []openai.ChatCompletionMessage{
	{Role: "assistant", Content: "Welcome to Mcdonalds, whenever you are ready, I would be happy to take your order."},
	{Role: "user", Content: "Yeah. Can I get a big mac and a large Sprite?"},
	{Role: "assistant", Content: "A Big Mac and a large Sprite. Would you like anything else with that?"},
	{Role: "user", Content: "Yeah. I'll get a small french fry."},
	{Role: "assistant", Content: "A Big Mac, small fries, and a large Sprite. Would you like anything else?"},
	{Role: "user", Content: "No, that will be all."},
	{Role: "assistant", Content: "A Big Mac, small fries, and a large Sprite is your final order. If everything looks good on screen, please proceed to the second window."},
}

func GetTextSimilarity(text1 string, text2 string) (float64, error) {
	embeddings, err := nomicai.EmbedText(
		os.Getenv("NOMIC_API_KEY"),
		[]string{text1, text2},
		nomicai.Clustering,
	)
	if err != nil {
		return 0, err
	}
	similarity, err := utils.CosineSimilarity(embeddings.Embeddings[0], embeddings.Embeddings[1])
	if err != nil {
		panic(err)
	}
	return similarity, nil
}

func ScriptComparison33() {
	godotenv.Load()
	llm_client := GetLLMClient(map[string]string{})

	// Gets the inital messages from the script
	pb := pocketbase.NewClient("https://example-you.pockethost.io")
	record, err := pb.GetRecord("templates", "uhf9k0x6kkcpx4x")
	if err != nil {
		panic(err)
	}
	llm_messages := utils.ScriptToConversation(record.Script)
	llm_messages = append(llm_messages, messages[0])

	// Loop through the messages and get the LLMs response
	for i, message := range messages {
		// Only continues if on a assistant message
		if message.Role == "assistant" {
			continue
		}
		// Adds the users message to the list of messages
		llm_messages = append(llm_messages, message)

		// Generates a new LLM Response
		llm_response, err := GetLLMResponse(llm_client, llm_messages, "openchat/openchat-7b")
		if err != nil {
			fmt.Println("Error fetching LLM response:", err)
			return
		}
		fmt.Println("User Message: " + messages[i].Content)
		fmt.Println("LLM Response: " + llm_response.Content)
		fmt.Println("Correct Response: " + messages[i+1].Content)

		// Adds the LLM response to the list of messages
		llm_messages = append(llm_messages, openai.ChatCompletionMessage{Role: "assistant", Content: llm_response.Content})
		embeddings, err := nomicai.EmbedText(
			os.Getenv("NOMIC_API_KEY"),
			[]string{
				messages[i+1].Content,
				llm_response.Content,
			},
			nomicai.Clustering,
		)
		if err != nil {
			panic(err)
		}

		similarity, err := utils.CosineSimilarity(embeddings.Embeddings[0], embeddings.Embeddings[1])
		if err != nil {
			panic(err)
		}
		fmt.Println(similarity)
		if similarity < 0.9 {
			fmt.Println("The LLM response is not similar enough to the correct response.")
			return
		}
	}
}
