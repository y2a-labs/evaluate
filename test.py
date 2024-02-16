import openai
import scripts
from pprint import pprint
llm_client = openai.Client(base_url='https://openrouter.ai/api/v1', api_key='d00e17c60e1e3138300de6cb201af9aa94f318252a5989ef553e16f97c109b4c')

condition = 'If the task is complete, say PASS then reply to the user.'
character = "Keep your responses short and to the point. You are are trying to complete the defined task. Say (IN PROGRESS) if the task is still in progress, and (DONE) if the task is complete at the beginning of each message then reply to the user."

def get_response(messages, ps, next_step):
    system_prompt_content = f"""CHARACTER:{character}
TASK: {ps['task']} {ps['condition']}"""

    messages[0] = {'role': 'system', 'content': system_prompt_content}
    messages.append({'role': 'assistant', 'content': '('})
    resp = llm_client.chat.completions.create(
        model='openchat/openchat-7b',
        messages=messages
        )
    
    return resp.choices[0].message.content



current_process = 0

messages = [
    {'role': 'system', 'content': ''},
    {'role': 'assistant', 'content': '(IN PROGRESS) ' + scripts.process_steps[current_process]['intro']},
    #{'role': 'user', 'content': 'Hello?'},
    #{'role': 'assistant', 'content': '(IN PROGRESS) ' + "It's Stacy from Elijah's Team. I'm reaching out to see how we can assist you further. How's your day going so far?"},
    #{'role': 'user', 'content': 'Pretty good, thanks.'},
    ]

print(messages[len(messages) - 1]['content'])

while True:
    if messages[len(messages) - 1]['role'] == 'assistant':
        user_input = input('You: ')
        messages.append({'role': 'user', 'content': user_input})

    resp = get_response(messages, scripts.process_steps[current_process], scripts.process_steps[current_process + 1])

    if '(DONE)' in resp:
        # Remove "PASS" from the response
        #resp = resp.replace('(DONE)', '').strip()
        # only keep the first sentence of the response
        #resp = resp.split('.')[0] + "."
        print('--- going to the next process! ---')
        current_process += 1
        if current_process >= len(scripts.process_steps):
            print('All processes are done!')
            break
        resp = f"{resp} {scripts.process_steps[current_process]['intro']}"
        # Resets the messages back to 0
        #messages = [
        #    {'role': 'system', 'content': ''},
        #    {'role': 'assistant', 'content': resp}
        #    ]
    elif 'FAIL' in resp:
        print('failed the process, try again')
        break
        #resp = process_steps[current_process]['intro']

    messages.append({'role': 'assistant', 'content': resp})
    print(resp)