#key = "nk-lqZpGiDLtDo7oHkFVlSNEvzTITGjsHiU3YaurgrtYDs"
import time
from nomic import embed
start_time = time.time()
embed.text(
    texts=['Nomic Embedding API'],
    model='nomic-embed-text-v1',
    task_type='clustering'
)

print(f"Duration: {time.time() - start_time}")