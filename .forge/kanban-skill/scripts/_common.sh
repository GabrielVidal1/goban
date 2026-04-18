#!/usr/bin/env bash
# _common.sh — sourced by all kanban scripts
# Provides: resolve_project, extract_field, find_ticket

KANBAN_DIR="${KANBAN_DIR:-./kanban}"

# resolve_project [--project <name>]
# Sets PROJECT_DIR. If --project is omitted and only one project exists, uses it.
# Populates PROJECT_NAME.
resolve_project() {
  local project_arg=""
  # Scan for --project in the remaining args (caller already parsed its own args)
  # We accept it as a pre-parsed variable: PROJECT_ARG
  project_arg="${PROJECT_ARG:-}"

  if [[ ! -d "$KANBAN_DIR" ]]; then
    echo "Error: kanban directory not found at '$KANBAN_DIR'" >&2
    exit 1
  fi

  # Collect project dirs (subdirs that contain at least one subdir themselves,
  # i.e. they look like project roots, not columns)
  mapfile -t PROJECTS < <(
    for d in "$KANBAN_DIR"/*/; do
      [[ -d "$d" ]] || continue
      bname="$(basename "$d")"
      [[ "$bname" == _* ]] && continue
      # A project dir contains subdirectories (columns)
      if find "$d" -mindepth 1 -maxdepth 1 -type d | grep -q .; then
        echo "$bname"
      fi
    done
  )

  if [[ -n "$project_arg" ]]; then
    PROJECT_DIR="$KANBAN_DIR/$project_arg"
    PROJECT_NAME="$project_arg"
    if [[ ! -d "$PROJECT_DIR" ]]; then
      echo "Error: project '$project_arg' not found in $KANBAN_DIR" >&2
      exit 1
    fi
  elif [[ ${#PROJECTS[@]} -eq 1 ]]; then
    PROJECT_NAME="${PROJECTS[0]}"
    PROJECT_DIR="$KANBAN_DIR/$PROJECT_NAME"
  elif [[ ${#PROJECTS[@]} -eq 0 ]]; then
    echo "Error: no projects found in $KANBAN_DIR" >&2
    echo "Expected structure: kanban/<project>/<column>/<ticket>.md" >&2
    exit 1
  else
    echo "Error: multiple projects found — specify one with --project <name>" >&2
    echo "Projects: ${PROJECTS[*]}" >&2
    exit 1
  fi
}

# list_projects — prints all project names
list_projects() {
  for d in "$KANBAN_DIR"/*/; do
    [[ -d "$d" ]] || continue
    bname="$(basename "$d")"
    [[ "$bname" == _* ]] && continue
    if find "$d" -mindepth 1 -maxdepth 1 -type d | grep -q .; then
      echo "$bname"
    fi
  done
}

# extract_field <file> <field>
extract_field() {
  local file="$1" field="$2"
  awk "
    /^---/ { if (fm++) exit }
    fm && /^${field}:/ {
      sub(/^${field}:[[:space:]]*/, \"\")
      gsub(/['\''\"]/,\"\")
      print
      exit
    }
  " "$file"
}

# find_ticket <project_dir> <slug>
# Prints the path of the matching ticket or exits with error.
find_ticket() {
  local proj_dir="$1" query="$2"
  if [[ -f "$query" ]]; then
    echo "$query"
    return
  fi
  local slug="${query%.md}"
  local match
  match=$(find "$proj_dir" -type f -name "*.md" | grep -i "${slug}" | head -1)
  if [[ -z "$match" ]]; then
    echo "Error: ticket '$query' not found in $proj_dir" >&2
    exit 1
  fi
  echo "$match"
}
