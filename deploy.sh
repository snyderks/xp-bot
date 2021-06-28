#! /bin/sh

mv ./xp-bot-dest xp-bot-prod
./xp-bot-prod >> log/out.txt 2>> log/log.txt &
