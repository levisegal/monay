#!/bin/bash
set -e

RALPH_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_ROOT="$RALPH_DIR/.."
LOG_DIR="$RALPH_DIR/logs"
mkdir -p "$LOG_DIR"

SESSION_ID=$(date +%Y%m%d_%H%M%S)
LOG_FILE="$LOG_DIR/ralph_$SESSION_ID.log"

echo "Ralph session: $SESSION_ID"
echo "Logging to: $LOG_FILE"

verify() {
    echo "=== Verification ===" | tee -a "$LOG_FILE"

    # Go backend
    cd "$REPO_ROOT/services/holdings"
    if ! make test 2>&1 | tee -a "$LOG_FILE"; then
        echo "FAIL: Go tests" | tee -a "$LOG_FILE"
        return 1
    fi

    # TypeScript frontend
    cd "$REPO_ROOT/services/client"
    if ! pnpm tsc:check 2>&1 | tee -a "$LOG_FILE"; then
        echo "FAIL: TypeScript check" | tee -a "$LOG_FILE"
        return 1
    fi

    if ! pnpm build 2>&1 | tee -a "$LOG_FILE"; then
        echo "FAIL: Next.js build" | tee -a "$LOG_FILE"
        return 1
    fi

    echo "PASS: All checks" | tee -a "$LOG_FILE"
    return 0
}

auto_commit() {
    cd "$REPO_ROOT"
    if git diff --quiet && git diff --cached --quiet; then
        echo "No changes to commit" | tee -a "$LOG_FILE"
        return 0
    fi

    # Get the last completed task for commit message
    TASK=$(grep -m1 "^\- \[x\]" "$RALPH_DIR/TASKS.md" | sed 's/- \[x\] //' | head -c 72)
    if [ -z "$TASK" ]; then
        TASK="ralph iteration $iteration"
    fi

    git add -A
    git commit -m "ralph: $TASK" | tee -a "$LOG_FILE"
    echo "Committed: ralph: $TASK" | tee -a "$LOG_FILE"
}

run_claude() {
    cd "$REPO_ROOT"
    cat "$RALPH_DIR/PROMPT.md" | claude --print 2>&1 | tee -a "$LOG_FILE"
}

iteration=0
max_iterations=${1:-10}  # Default 10, or pass as arg

while [ $iteration -lt $max_iterations ]; do
    iteration=$((iteration + 1))
    echo ""
    echo "========================================" | tee -a "$LOG_FILE"
    echo "Iteration $iteration of $max_iterations" | tee -a "$LOG_FILE"
    echo "========================================" | tee -a "$LOG_FILE"

    # Check if tasks remain
    if ! grep -q "^\- \[ \]" "$RALPH_DIR/TASKS.md" 2>/dev/null; then
        echo "No pending tasks. Ralph is done!" | tee -a "$LOG_FILE"
        break
    fi

    # Run Claude
    run_claude

    # Verify and commit if successful
    if verify; then
        echo "Iteration $iteration: SUCCESS" | tee -a "$LOG_FILE"
        auto_commit
    else
        echo "Iteration $iteration: FAILED verification" | tee -a "$LOG_FILE"
        echo "Ralph will try to fix on next iteration..."
    fi

    # Brief pause to avoid rate limits
    sleep 2
done

echo ""
echo "Ralph completed $iteration iterations"
echo "Log: $LOG_FILE"
