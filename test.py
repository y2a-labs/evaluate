import requests

# Your API key
api_key = 'sk-helicone-bepv2zy-as3ebba-vwdihja-ocbs4sa'

# Headers
headers = {
    'Authorization': f'Bearer {api_key}',
    'Content-Type': 'application/json',
}

# JSON data to be sent in the POST request
data = {
    "providerRequest": {
        "url": "https://example.com",
        "json": {
            "key1": "value1",
            "key2": "value2",
            "model": "openchat"
        },
        "meta": {
            "metaKey1": "metaValue1",
            "metaKey2": "metaValue2"
        }
    },
    "providerResponse": {
        "json": {
            "responseKey1": "responseValue1",
            "responseKey2": "responseValue2",
            "first_token_latency_ms": 500,
            "first_sentence_latency_ms": 1000,
        },
        "status": 200,
        "headers": {
            "Helicone-Property-cid": "4353",
        }
    },
    "timing": {
        "startTime": {
            "seconds": 1709101100,
            "milliseconds": 500
        },
        "endTime": {
            "seconds": 1709101104,
            "milliseconds": 750
        }
    }
}

# The URL to which the request is sent
url = 'https://api.hconeai.com/oai/v1/log'

# Sending the POST request
response = requests.post(url, json=data, headers=headers)

# Printing the response from the server
print(response.text)