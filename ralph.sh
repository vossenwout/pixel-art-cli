set -e

if [ -z "$1" ]; then
  echo "Usage: $0 <iterations>"
  exit 1
fi

prompt=$(
  cat <<'EOF'
@prd.md @tasks.md @progress.md
1. Find the highest-priority feature in tasks.md to work on and work only on that feature.
This should be the one YOU decide has the highest priority - not necessarily the first in the list
2. Write tests for the acceptance criteria of the feature that are noted in tasks.md and check that tests pass before you decide the criteria is done
3. If you managed to finish acceptance criteria than update tasks.md checking of the acceptance criteria by putting an x in the markdow checkbox.
4. Use the progress.md file to document your progress.
Use this to leave a note for the next person working in the codebase.
However don't leave useless bloat that confuses that person.
5. Make a git commit of the feature with the feature you worked on in the name. However don't push to origin.
  6. If you notice that tasks.md is complete (all the acceptance criteria of all features are complete) (all the acceptance criteria of all features are complete) (all the acceptance criteria of all features are complete) (all the acceptance criteria of all features are complete) (all the acceptance criteria of all features are complete) (all the acceptance criteria of all features are complete) (all the acceptance criteria of all features are complete) (all the acceptance criteria of all features are complete) (all the acceptance criteria of all features are complete), output <promise>COMPLETE</promise>
EOF
)

for ((i = 1; i <= $1; i++)); do
  echo "Iteration $i"
  echo "-------------"
  result=$(codex exec --config model_reasoning_effort="high" --yolo "${prompt}")
  echo "$result"
  if [[ "$result" == *"<promise>COMPLETE</promise>"* ]]; then
    echo "PRD complete, exiting."
    exit 0
  fi
done
