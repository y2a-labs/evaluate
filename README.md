# Y2A Multi-Turn Conversation Evaluator

Discover your ideal AI model. It provides a GUI for managing various language learning model (LLM) API providers, logs all your requests in a SQLite database, and allows you to compare and score model outputs against each other. This enables you to find faster and more cost-effective models to use.
![image](https://github.com/y2a-labs/evaluate/assets/151597434/492d3e6b-0a8b-4854-b201-3abad6171540)

## Key Features

- **LLM Proxy**: Manage all your LLM API providers securely from a single source. This includes providers like OpenAI, Together, OpenRouter, Local, and more.
- **Logging**: Keep track of all your requests with a robust SQLite database logging system.
- **Testing**: Compare the outputs of different models and score the results against each other. This feature helps you identify faster and cheaper models.

## Getting Started

Follow these steps to get started with Y2A Evaluator:

1. **Install Go**: Ensure you have Go version 1.22 or higher installed on your system. If not, you can install it from the [official Go website](https://go.dev/doc/install).
2. **Install the Evaluator CLI tool**: Use the following command to install the Evaluator CLI tool:
    ```bash
    go install github.com/y2a-labs/evaluate@latest
    ```
3. **Start the server**: Run the following command to start the server:
    ```bash
    evaluate server
    ```
4. **Add your API providers**: OpenAI is required for text embedding, but all other providers are optional.
5. **Log your requests**: Update your base url and set the model name to any of the providers
    ```python
    from openai import OpenAI

    client = OpenAI(
    base_url="http://localhost:3000/v1/",
    api_key="any",
    )

    response = client.chat.completions.create(
    model="Your model name"
    messages=[
    {"role": "user","content": "How are you doing today?"}
    ])

    print(response.choices[0].message.content)
    ```
6. **Create a test**: Convert a log of a previous request into a test, or make one from scratch.

## Community

Got any questions? Join our [Discord Community!](https://discord.gg/HXgSS7RuWc).
