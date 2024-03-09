import openai

client = openai.Client(
    base_url="http://127.0.0.1:3000/v1",
    api_key="sdf",
    default_headers={"AGENT-ID": "8d9fc551-5e34-411f-af91-52d8ab8e28dc"},
)

stream = client.chat.completions.create(
    model="gpt-3.5-turbo",
    stream=True,
    messages=[
        {
            "role": "system",
            "content": "You are a helpful assistant."
        },
        {
            "role": "user",
            "content": "What is the meaning of life?"
        }
    ]
)

for chunk in stream:
    print(chunk)