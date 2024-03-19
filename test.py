from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:3000/v1/",
    api_key="any"
)

response = client.embeddings.create(
    input="Hello, world",
    model="text-embedding-3-small"
)

print(response)