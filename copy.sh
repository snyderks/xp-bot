#! /bin/sh

scp ./xp-bot-prod xp-bot:~/xp-bot-dest
scp ./colors.toml xp-bot:~/colors.toml
scp ./deploy.sh xp-bot:~/deploy.sh