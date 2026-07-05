# Recorded with vhs (github.com/charmbracelet/vhs):
#   vhs demos/list/demo.tape
Output demos/list/demo.gif

Set FontSize 18
Set Width 860
Set Height 560
Set Padding 12
Set TypingSpeed 120ms

Hide
Type "go run ./demos/list"
Enter
Sleep 3s
Show

Sleep 1s
Down
Sleep 500ms
Down
Sleep 500ms
Down
Sleep 800ms

# Open the confirm modal — focus is trapped inside it
Enter
Sleep 1.2s
Tab
Sleep 700ms
Tab
Sleep 700ms

# Dismiss with Escape, focus restored
Escape
Sleep 1s

Up
Sleep 500ms

# Confirm this time
Enter
Sleep 1s
Enter
Sleep 1.2s

Type "q"
Sleep 1s
