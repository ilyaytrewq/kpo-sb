docker exec -it broker /opt/kafka/bin/kafka-topics.sh --bootstrap-server broker:9092 \
  --create --if-not-exists \
  --topic payments.payment_requested.v1 \
  --partitions 3 --replication-factor 1

docker exec -it broker /opt/kafka/bin/kafka-topics.sh --bootstrap-server broker:9092 \
  --create --if-not-exists \
  --topic payments.payment_result.v1 \
  --partitions 3 --replication-factor 1
