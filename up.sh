go run main.go &&
chmod 0777 ./app &&
cd ./app &&
docker-compose up -d --build &&
docker-compose ps &&
pwd &&
cp ../node_modules.tar ../app/front/node_modules.tar &&
cd front &&
tar -xvf node_modules.tar &&
rm node_modules.tar &&
cd ../ &&
echo "wait 15 sec for Database"
sleep 15 &&
docker exec -i app_mysql_1 mysql -uroot -proot < sql.sql &&
cd back &&
go get -d -v ./... && go install -v ./... && go build -o server .
echo "&& (./server &) && cd ../front && (npm start &)"
