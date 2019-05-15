import json
from collections import defaultdict
import psycopg2
import sys

def load_country_tuples():
    with open("countries_en.json") as f:
        en = json.load(f)

    with open("countries_ru.json") as f:
        ru = json.load(f)

    uni = defaultdict(lambda: [None]*2)

    for c in ru:
        uni[c[0]][0] = c[1]
    for c in en:
        uni[c[0]][1] = c[1]

    for code, names in uni.items():
        yield code, names[0], names[1]

# Open connection

if len(sys.argv) != 2:
    print('usage: {} pg-conn'.format(sys.argv[0]))
    exit(1)

conn = psycopg2.connect(sys.argv[1])
cur = conn.cursor()

# Initialize Schema
with open('kino.sql') as f:
    cur.execute(f.read())

# Fill Timezones
cur.execute('''
INSERT INTO timezones (name, utc_offset)
SELECT name, utc_offset
FROM pg_timezone_names
ON CONFLICT ON CONSTRAINT timezones_name_key DO NOTHING
''')

# Fill countries
cur.execute('''
INSERT INTO countries (country_code, name_ru, name_en)
VALUES
{}
ON CONFLICT (country_code) DO NOTHING
'''.format(
    ', '.join(
        cur.mogrify("(%s,%s,%s)", c).decode() for c in load_country_tuples())
)
)

# Close connection gracefully
conn.commit()
cur.close()
conn.close()
