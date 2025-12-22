cd embedding-service
docker compose up -d

cd ..
cd file-storing
docker compose up -d


cd ..
cd file-analisys
docker compose up -d


cd ..
cd api-gateway
docker compose up -d