RPMBUILD_PATH='/root/rpmbuild/'
ROOT_PATH="$(pwd)"
OS_RELEASE="$(cat /etc/centos-release)"

echo "$ROOT_PATH"

#remove build folder
if [[ -d "$RPMBUILD_PATH" ]]; then
    rm -rf "$RPMBUILD_PATH"
fi

#create build folder
mkdir -p $RPMBUILD_PATH/{BUILD,BUILDROOT,RPMS,SOURCES,SPECS,SRPMS}

#checkout latest commit or release tag

#build sources
RET="$($ROOT_PATH/build.sh)"

# echo $RET

if [[ "$RET" == "failed" ]]; then
    echo "Something went wrong when run ./build.sh"
    exit 1
fi

if [[ -d "$ROOT_PATH/TELEBOT-0.0.0" ]]; then
    rm -rf $ROOT_PATH/TELEBOT-0.0.0
fi

#tar binary -> gz file
cp -R $ROOT_PATH/bin $ROOT_PATH/TELEBOT-0.0.0
tar -czvf TELEBOT-0.0.0.tar.gz TELEBOT-0.0.0/

cp $ROOT_PATH/TELEBOT-0.0.0.tar.gz $RPMBUILD_PATH/SOURCES
rm -rf $ROOT_PATH/TELEBOT-0.0.0
rm -rf $ROOT_PATH/TELEBOT-0.0.0.tar.gz

if [[ $? -ne 0 ]]; then
    echo "Somthing went wrong when run tar command"
    exit 1
fi

#copy sources to SOURCES and spec file to SPECS
# cd to root, exec rpmbuild command
cd /root
if [[ $OS_RELEASE == *"release 7."* ]]; then
    cp $ROOT_PATH/src/bot/telebot_centos7.spec $RPMBUILD_PATH/SPECS
    rpmbuild --clean -bb rpmbuild/SPECS/telebot_centos7.spec
else
    cp $ROOT_PATH/src/bot/telebot_centos6.spec $RPMBUILD_PATH/SPECS
    rpmbuild --clean -bb rpmbuild/SPECS/telebot_centos6.spec
fi
# echo $(pwd)

