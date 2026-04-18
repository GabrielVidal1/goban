#!/usr/bin/env bash
# script.sh — Process a ticket using muse agent to plan it, then move to To Do.
# Usage: ./script.sh <ticket-slug>

set -euo pipefail

if [[ $# -lt 1 ]]; then
  echo "Usage: $0 <ticket-slug>" >&2
  exit 1
fi

TICKET_SLUG="$1"

# Resolve the kanban directory (same as _common.sh)
KANBAN_DIR="${KANBAN_DIR:-./kanban}"
PROJECT="kanban-ui"
COLUMN="Backlog"

PROJ_DIR="$KANBAN_DIR/$PROJECT"

# Find the ticket file
TICKET_FILE=$(find "$PROJ_DIR" -type f -name "*.md" | grep -i "${TICKET_SLUG}" | head -1)
if [[ -z "$TICKET_FILE" ]]; then
  echo "Error: ticket '$TICKET_SLUG' not found in $PROJ_DIR" >&2
  exit 1
fi

# Read the full ticket content (title + body)
TITLE=$(awk '/^---/{fm++;next} fm && /^---/{exit} fm{print}' "$TICKET_FILE")
BODY=$(sed -n '/^---$/,/^---$/{ /^---$/d;p; }' "$TICKET_FILE" | tail -n +2)

echo "=== Processing ticket: $TITLE ==="
echo ""

# Build the prompt for muse agent
PROMPT="You are planning a task from this kanban ticket. Analyze it carefully and create a detailed implementation plan.

Ticket Title: $TITLE
Ticket Body: $BODY

Please provide:
1. A clear understanding of what needs to be done
2. Step-by-step implementation plan
3. Any dependencies or considerations
4. Verification criteria

Be thorough but practical."

# Run muse agent via forge with the ticket content as prompt
echo "Running muse agent to plan this ticket..."
echo ""
forge --agent muse -p "$PROMPT" -C "$KANBAN_DIR/$PROJECT"

echo ""
echo "=== Plan complete ==="
echo ""

# Move the ticket to To Do column using the kanban script
SCRIPTS_DIR="$(dirname "$(readlink -f "$0")")/../scripts"
if [[ ! -d "$SCRIPTS_DIR" ]]; then
  # Fallback: use the skill scripts location
  SCRIPTS_DIR="/Users/gabrielvidal/forge/skills/kanban-skill/scripts"
fi

bash "$SCRIPTS_DIR/update-ticket-status" "$TICKET_SLUG" "To Do" --project "$PROJECT"

echo "Ticket '$TITLE' moved to To Do column."
