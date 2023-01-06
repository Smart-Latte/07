import sqlite3

dbname = 'db/20230106.db'
conn = sqlite3.connect(dbname)
cur = conn.cursor()

cur.execute('SELECT * FROM EnergyData ORDER BY GeneratedTime LIMIT 10')
for row in cur:
    print(row)

cur.execute('SELECT * FROM BidData ORDER BY Amount DESC LIMIT 10')
for row in cur:
    print(row)

cur.execute("SELECT LargeCategory, total(Amount), total(SoldAmount) FROM EnergyData GROUP BY LargeCategory")
for row in cur:
    print(row)

cur.close()

conn.close()