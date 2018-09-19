#!/bin/bash


while getopts ":s:l:" opt; do
  case $opt in
    s)
      echo "-s was triggered, Parameter: $OPTARG" >&2
      SERVICE=$OPTARG
      ;;
    l)
      echo "launching service... $OPTARG" >&2
      SERVICE=$OPTARG
      ;;
    \?)
      echo "Invalid option: -$OPTARG" >&2
      exit 1
      ;;
    :)
      echo "Option -$OPTARG requires an argument." >&2
      exit 1
      ;;
  esac
done

echo "$SERVICE" >&2