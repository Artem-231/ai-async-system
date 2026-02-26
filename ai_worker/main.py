import pika
import psycopg2
import os
import json
import requests
from PIL import Image
from io import BytesIO

DB_USER = os.getenv("DB_USER")
DB_PASSWORD = os.getenv("DB_PASSWORD")
HF_TOKEN = os.getenv("HF_TOKEN")

OUTPUT_DIR = "/app/images"
if not os.path.exists(OUTPUT_DIR):
    os.makedirs(OUTPUT_DIR)

API_URL = "https://router.huggingface.co/hf-inference/models/stabilityai/stable-diffusion-xl-base-1.0"
headers = {"Authorization": f"Bearer {HF_TOKEN}"}

def generate_image(prompt, task_id):
    print(f"Начинаю генерацию для задачи {task_id}: {prompt}")
    try:
        response = requests.post(API_URL, headers=headers, json={"inputs": prompt})

        if response.status_code != 200:
            print(f"Ошибка API: {response.text}")
            return False

        image = Image.open(BytesIO(response.content))
        image_path = f"/app/images/{task_id}.png"
        image.save(image_path)

        print(f"Задача {task_id} успешно завершена!")
        return True

    except Exception as e:
        print(f"Критическая ошибка при генерации: {e}")
        return False

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
