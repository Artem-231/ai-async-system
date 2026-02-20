import os
import time
import json
import pika
import psycopg2
import requests

HF_TOKEN = os.getenv("HF_TOKEN")
MODEL_URL = os.getenv("MODEL_URL")
HEADERS = {"Authorization": f"Bearer {HF_TOKEN}"}

DB_USER = os.getenv("DB_USER")
DB_PASSWORD = os.getenv("DB_PASSWORD")

OUTPUT_DIR = "/app/images"
if not os.path.exists(OUTPUT_DIR):
    os.makedirs(OUTPUT_DIR)

# Функция генерации
def generate_image(prompt, task_id):

    print(f" [AI] Рисую: {prompt}...")

    payload = {"inputs": prompt}

    response = requests.post(MODEL_URL, headers=HEADERS, json=payload)

    if response.status_code != 200:
        print(f" [AI] Ошибка API: {response.text}")
        return None

    file_name = f"{task_id}.png"
    file_path = os.path.join(OUTPUT_DIR, file_name)

    with open(file_path, "wb") as f:
        f.write(response.content)

    print(f" [AI] Картинка сохранена: {file_path}")
    return file_name

# Функция для работы с БД
def update_task_status(task_id, status):
    try:
        conn = psycopg2.connect(
            dbname="aiAsyncSystem",
            user=DB_USER,
            password=DB_PASSWORD,
            host="postgres",
            port="5432"
        )
        cursor = conn.cursor()
        cursor.execute("UPDATE tasks SET status = %s WHERE id = %s", (status, task_id))
        conn.commit()
        cursor.close()
        conn.close()
        print(f" Задача {task_id} обновлена в БД -> {status}")
    except Exception as e:
        print(f" Ошибка БД: {e}")

def callback(ch, method, properties, body):
    try:
        data = json.loads(body)
        task_id = data.get("id")
        prompt = data.get("payload", "cat")

        print(f" [x] Получена задача {task_id}: {prompt}")

        image_file = generate_image(prompt, task_id)

        if image_file:
            update_task_status(task_id, "done")
        else:
            update_task_status(task_id, "error")

        ch.basic_ack(delivery_tag=method.delivery_tag)
        print(" [x] Done (Ack sent)")

    except Exception as e:
        print(f"Error processing message: {e}")
        ch.basic_nack(delivery_tag=method.delivery_tag)

def main():
    try:
        connection = pika.BlockingConnection(
            pika.ConnectionParameters(host='rabbitmq', port=5672)
        )
        channel = connection.channel()
        channel.queue_declare(queue='task_queue', durable=True)

        print(' [*] Waiting for messages. To exit press CTRL+C')

        channel.basic_qos(prefetch_count=1)
        channel.basic_consume(queue='task_queue', on_message_callback=callback)
        channel.start_consuming()

    except Exception as e:
        print(f"Failed to connect to RabbitMQ: {e}")

if __name__ == '__main__':
    main()
