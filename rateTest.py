import requests
import os
from pprint import pprint
OPENROUTER_API_KEY = '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'
print(OPENROUTER_API_KEY)
headers = {
    'Authorization': f'Bearer {OPENROUTER_API_KEY}',
}

response = requests.get('https://openrouter.ai/api/v1/auth/key', headers=headers)

pprint(response.json())