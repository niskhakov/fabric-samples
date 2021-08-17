import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import json

patternToFind = "State:{"

metrics = []

entriesDict = {}

for i in range(1, 11):
    with open("stress{}.log".format(i), 'r') as reader:
        lines = reader.readlines()
        empty_lines = 0
        for line in lines:
            jsonResp = line[line.index(patternToFind) + len(patternToFind) - 1:-2]
            jsonResp = jsonResp.replace(" \"", "")
            jsonResp = jsonResp.replace("\\", "")
            # print(jsonResp)
            if jsonResp == "":
                empty_lines += 1
                break
            obj = json.loads(jsonResp)
            metrics.append(obj)
            print(obj)
        print("Empty lines: {}".format(empty_lines))

# exit(0)
# Get only put operations
putEntries = filter(lambda x: x["method"] == 'put' and x["batchapi"] == True, metrics)
putEntriesNoBatch = filter(lambda x: x["method"] == 'put' and x["batchapi"] == False, metrics)
x = []
y = []
for entry in putEntries:
    x.append(entry["entries"])
    y.append(entry["millis"])
    print(entry)
xn = []
yn = []
for entry in putEntriesNoBatch:
    xn.append(entry["entries"])
    yn.append(entry["millis"])
    print(entry)

plt.title("BatchAPI vs Standard Put")
plt.grid(True)
plt.xlabel("Number of keys")
plt.ylabel("Milliseconds")
plt.plot(x, y, 'ro', label="BatchAPI")
plt.plot(xn, yn, 'go', label="Standard")
plt.legend()
plt.show()