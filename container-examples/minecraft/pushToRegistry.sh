docker build . -t minecraft
docker tag minecraft localhost:5000/minecraft
docker push localhost:5000/minecraft

