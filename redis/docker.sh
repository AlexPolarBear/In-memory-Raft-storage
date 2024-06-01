#!/bin/bash

echo "Сборка образа Redis..."
docker build -t my_redis_image .

echo "Запуск контейнера Redis..."
docker run -d -p6379:6379 --rm my_redis_image

exit0
