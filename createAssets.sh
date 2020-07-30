

echo "creating tables"



aws dynamodb create-table \
    --table-name OCRTableProcess \
    --attribute-definitions \
        AttributeName=FileHash,AttributeType=S \
    --key-schema AttributeName=FileHash,KeyType=HASH \
    --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1

    