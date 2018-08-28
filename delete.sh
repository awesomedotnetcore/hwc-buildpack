cf apps | awk '/windows_app_/ {print $1}' | while read app; do cf delete -f $app; done
