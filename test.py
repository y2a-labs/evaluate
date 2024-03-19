from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:3000/v1/",
    api_key="any"
)

response = client.chat.completions.create(
  model="gpt-3.5-turbo",
  messages=[
    {
      "role": "user",
      "content": "How are you?"
    }
  ],
  stream=True
)

for chunk in response:
    print(chunk)