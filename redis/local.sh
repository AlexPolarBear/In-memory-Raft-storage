#!/bin/bash


# Установка Redis на каждый сервер
sudo apt-get update && sudo apt-get upgrade
sudo apt install make gcc libc6-dev tcl

wget http://download.redis.io/redis-stable.tar.gz
tar xvzf redis-stable.tar.gz
cd redis-stable
sudo make install

make test


# Настройка узлов Master и Slave
USERNAME=root
SERVER_IP=127.0.0.1

setup_master() {
    echo "Подключаемся к серверу \$1 по SSH для настройки master..."
    ssh $USERNAME@\$1 <<EOF
    echo "Находимся на сервере \$1. Начинаем настройку master..."
    cd redis-stable/
    cp redis.conf \$2_master.conf
    sed -i 's/^bind .*/bind 127.0.0.1/' \$2_master.conf
    sed -i 's/^protected-mode .*/protected-mode no/' \$2_master.conf
    sed -i "s/^port .*/port \$2/" \$2_master.conf
    sed -i "s/^pidfile .*/pidfile \/var\/run\/redis_\$2.pid/" \$2_master.conf
    sed -i 's/^cluster-enabled .*/cluster-enabled yes/' \$2_master.conf
    sed -i "s/^cluster-config-file .*/cluster-config-file nodes-\$2.conf/" \$2_master.conf
    sed -i 's/^cluster-node-timeout .*/cluster-node-timeout 15000/' \$2_master.conf
EOF
}

setup_slave() {
    echo "Подключаемся к серверу \$1 по SSH для настройки slave..."
    ssh $USERNAME@\$1 <<EOF
    echo "Находимся на сервере \$1. Начинаем настройку slave..."
    cd redis-stable/
    cp redis.conf \$2_slave.conf
    sed -i 's/^bind .*/bind 127.0.0.1/' \$2_slave.conf
    sed -i 's/^protected-mode .*/protected-mode no/' \$2_slave.conf
    sed -i "s/^port .*/port \$2/" \$2_slave.conf
    sed -i "s/^pidfile .*/pidfile \/var\/run\/redis_\$2.pid/" \$2_slave.conf
    sed -i 's/^cluster-enabled .*/cluster-enabled yes/' \$2_slave.conf
    sed -i "s/^cluster-config-file .*/cluster-config-file nodes-\$2.conf/" \$2_slave.conf
    sed -i 's/^cluster-node-timeout .*/cluster-node-timeout 15000/' \$2_slave.conf
EOF
}

setup_master $SERVER_IP 6379
setup_slave $SERVER_IP 6381

setup_master $SERVER_IP 6380
setup_slave $SERVER_IP 6379

setup_master $SERVER_IP 6381
setup_slave $SERVER_IP 6380


# Запуск узлов Master и Slave 
run_redis() {
    echo "Подключаемся к серверу \$1 по SSH..."
    ssh $USERNAME@\$1 <<EOF
    echo "Запускаем экземпляры Redis на сервере \$1..."
    redis-server redis-stable/\$2_master.conf
    redis-server redis-stable/\$3_slave.conf
EOF
}

run_redis $SERVER_IP 6379 6381
run_redis $SERVER_IP 6380 6379
run_redis $SERVER_IP 6381 6380


# Распределение данных
run_redis_commands() {
    echo "Подключаемся к Redis на сервере \$1..."
    redis-cli -c -h \$1 -p \$2 <<EOF
    CLUSTER INFO
    INFO replication
    SET John Adams
    SET James Madison
    SET Andrew Jackson
    GET John
EOF
}

run_redis_commands $SERVER_IP 6379
