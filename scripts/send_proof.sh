#!/bin/bash

# Take two arguments + optional third one: <plonk|sp1|cairo|kimchi> <proof_path> <pub_input_path?>
# If the third argument is not provided, the task sender will not send the public input
# The task sender will send the proof to the verifier with the given ID
if [[ "$#" -lt 2 || "$#" -gt 3 ]]; then
	echo "Usage: $0 <plonk|sp1|cairo|kimchi> <proof_path> <pub_input_path?>"
	exit 1
fi

# Run the task sender
if [[ "$#" -eq 2 ]]; then
	go run task_sender/cmd/main.go --verifier-id $1 --proof $2 2>&1 | zap-pretty
else 
	go run task_sender/cmd/main.go --verifier-id $1 --proof $2 --pub-input $3 2>&1 | zap-pretty
fi
