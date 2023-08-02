#!/bin/bash

# Function to read name and description from README.md
get_name_and_description() {
  local readme_file="$1"
  local name=""
  local description=""
  local in_description=false

  while IFS= read -r line; do
    if [[ "$line" =~ ^#\ ([A-Za-z0-9_\-]+)$ ]]; then
      name="${BASH_REMATCH[1]}"
    elif [[ "$line" =~ ^###\ Description$ ]]; then
      in_description=true
    elif [[ "$line" =~ ^###\ Usage$ ]]; then
      in_description=false
      break
    elif [[ "$in_description" == true ]]; then
      description+=" $line"
    fi
  done < "$readme_file"

  name=$(echo "$name" | tr '[:upper:]' '[:lower:]') # Convert name to lowercase
  description=$(echo "$description" | sed 's/^[[:space:]]*//') # Remove leading whitespace

  echo "$name" "$description"
}

# Main script
main() {
  top_level_readme="Index.md"
  echo "# Mods Index" > "$top_level_readme"
  echo "" >> "$top_level_readme"

  # Start the table
  echo "| Name | Description |" >> "$top_level_readme"
  echo "|------|-------------|" >> "$top_level_readme"

  # Loop through the immediate subdirectories (top-level folders) only
  for folder in */; do
    if [[ -d "$folder" ]]; then
      readme_file=$folder
      readme_file+="README.md"
      if [[ -f "$readme_file" ]]; then
        name_desc=$(get_name_and_description "$readme_file")
        name=$(echo "$name_desc" | awk '{print $1}')
        description=$(echo "$name_desc" | awk '{$1=""; print $0}')
        echo "| [$name]($readme_file) | $description |" >> "$top_level_readme"
      fi
    fi
  done
}

main