#!/bin/bash

source .env

SESSION_NAME="development"

# Create a new Tmux session
tmux -u new-session -d -s $SESSION_NAME

#This script requires autojump to be installed and alias added as 'j'

# Create windows with specified names and execute commands in each window
tmux -u new-window -t $SESSION_NAME:1 -n "nvim"
tmux send-keys -t $SESSION_NAME:1 'nvim' C-m

tmux -u new-window -t $SESSION_NAME:2 -n "ngrok"
tmux send-keys -t $SESSION_NAME:2 'ngrok http ${BOT_PORT}' C-m

tmux -u new-window -t $SESSION_NAME:3 -n "run"
tmux send-keys -t $SESSION_NAME:3 'sleep 5 && source genURL && air' C-m

tmux -u new-window -t $SESSION_NAME:4 -n "Lazygit"
tmux send-keys -t $SESSION_NAME:4 'lazygit' C-m

# Attach to the newly created session
tmux -u attach-session -t $SESSION_NAME
