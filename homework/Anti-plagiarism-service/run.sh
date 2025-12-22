cd embedding-service
docker compose up --build -d

cd ..
cd file-storing
docker compose up --build -d


cd ..
cd file-analisys
docker compose up --build -d


cd ..
cd api-gateway
docker compose up --build -d