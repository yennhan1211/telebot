GOPATH="$(pwd)"
export GOPATH=$GOPATH

rm -rf $GOPATH/bin/
mkdir -p $GOPATH/bin/

go build -o $GOPATH/bin/telebot bot

# cd bin/
# ./telebot.sh start

if [ $? -eq 0 ]; then
    cp $GOPATH/src/bot/telebot.sh $GOPATH/bin
    cp $GOPATH/src/bot/telebot.service $GOPATH/bin

    cp YOURPUBLIC.pem $GOPATH/bin/
    cp YOURPRIVATE.key  $GOPATH/bin/
    echo "success"
else
    echo "failed"
fi