
Name: TELEBOT
Version: 0.0.0
Release: 0%{?dist}
Summary: A bot for telegram to check crypto currency's price
License: MIT
URL: http://code4food.net

Source0: TELEBOT-0.0.0.tar.gz

AutoReqProv: no

%description
A bot for telegram to check crypto currency's price

%prep

%setup -q

%build

%install
rm -rf $RPM_BUILD_ROOT
mkdir -p $RPM_BUILD_ROOT/usr/local/twofive/telebot
mkdir -p $RPM_BUILD_ROOT/lib/systemd/system

cp YOURPUBLIC.pem  $RPM_BUILD_ROOT/usr/local/twofive/telebot/YOURPUBLIC.pem
cp YOURPRIVATE.key $RPM_BUILD_ROOT/usr/local/twofive/telebot/YOURPRIVATE.key
cp telebot.service $RPM_BUILD_ROOT/lib/systemd/system/telebot.service
cp telebot.service $RPM_BUILD_ROOT/usr/local/twofive/telebot/telebot.service
install -m 755 telebot.sh  $RPM_BUILD_ROOT/usr/local/twofive/telebot/telebot.sh
install -m 755 telebot $RPM_BUILD_ROOT/usr/local/twofive/telebot

%files
/usr/local/twofive/telebot/YOURPRIVATE.key
/usr/local/twofive/telebot/YOURPUBLIC.pem
/usr/local/twofive/telebot/telebot.sh
/usr/local/twofive/telebot/telebot
/usr/local/twofive/telebot/telebot.service
/lib/systemd/system/telebot.service

%changelog

