#!/usr/bin/env python
# coding: utf-8

# In[1]:


import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import json


# In[10]:


stressTests = []

patternToFind = "State:{"

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
            
            stressTests.append(obj)
            print(obj)
        print("Empty lines: {}".format(empty_lines))


# In[11]:


df = pd.DataFrame.from_records(stressTests)


# In[12]:


df


# In[28]:


meanDf = df.groupby(['entries']).mean()


# In[31]:


meanDf


# In[29]:


stdDf = df.groupby(['entries']).std()


# In[35]:


stdDf


# In[50]:


mean = np.array(meanDf['millis'])


# In[51]:


std = np.array(stdDf['millis'])


# In[59]:


entries = [name for name, _ in df.groupby(['entries'])]


# In[60]:


entries


# In[65]:

plt.title("BatchAPI Mean with Std")
plt.xlabel("Number of keys")
plt.ylabel("Milliseconds")
plt.errorbar(entries, mean, std, fmt='--o')
plt.show()
