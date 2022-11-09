# tombscripter

This is a pre-alpha library and command-line application for parsing and manipulating Tomb Raider script files (or "gameflow"), written in **Go**. It is fast (enough), dependency-free and runs anywhere you can get a Go runtime. It has support for the following games:
- Tomb Raider II (PC) 
- Tomb Raider III (PC) 

At the moment the program just dumps out a list of all levels in the script, and the gameflow that occurs for each level, pretty-printed. Not very useful but at the moment I am just testing the library.

Sample output:
```
Level 1: Jungle (data\jungle.TR2)
  Puzzles: [P1 P2 P3 P4]
  Keys: [K1 K2 K3 Indra Key]
  Pickups: [P1 P2]
  Flow: 
    0: Play Soundtrack 34
    1: Load Pic 'pix\india.bmp'
    2: Play Level 'data\jungle.TR2' (Jungle)
    3: Play Soundtrack 64
    4: Set Cutscene Angle 16384
    5: Display Cutscene 'cuts\cut6.TR2'
    6: End Level
    7: End Sequence
```

## Why? 
I needed something like this for a project. Other such programs exist, but they are closed-source, Windows only and suck [citation needed].
