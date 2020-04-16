import pandas
import matplotlib.pyplot as plot
import matplotlib.dates as dates
from datetime import datetime

def parse_label(label):
    x = [v.split('="') for v in label[1:-1].split(",")]
    return dict([[v[0], v[1][:-1]] for v in x])

d = pandas.read_csv('mem.csv')
time = [datetime.utcfromtimestamp(x) for x in d['Time']]

columns = dict([(x, parse_label(x)) for x in d.columns if x != 'Time'])

fig = plot.figure(figsize = (12, 10))
ax = fig.add_subplot(111)

col = d.columns[1]
data = [x / 1000000 for x in d[col]]
ax.plot(time, data, label = "Used", color = "tomato")

ax.set_title("Memory Usage")
ax.set_xlabel("Time (UTC)")
ax.set_ylabel("Used (MBytes)")
ax.set_xlim(time[0], time[-2])
ax.set_ylim(0)
ax.xaxis.set_major_locator(dates.HourLocator(interval = 3))
ax.xaxis.set_major_formatter(dates.DateFormatter("%m/%d %H:%M"))
ax.grid(True)
ax.legend()

plot.setp(ax.get_xticklabels(), rotation = 45)
plot.savefig("mem.svg")
