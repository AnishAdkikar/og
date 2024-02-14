import mysql.connector as mysql

# Connect to MySQL database
my_database = mysql.connect(host="localhost", 
                          user="root",
                          password="1qwerty9",
                          database="nodal_officers")
# Create cursor 
my_cursor = my_database.cursor()

# Create cursor  object with dictionary factory enabled so that we can use column names instead of column indices
# my_cursor=my_database.cursor(dictionary=True)

query = "SELECT * FROM list_nodal_officer WHERE State='Karnataka'"

# Execute query
my_cursor.execute(query)

for row in my_cursor:
    # Converting row to string from tuple
    result = str(row)
    # Removing brackets from the string
    result = result.replace('(','').replace(')','')
    print(result)
    
    

my_cursor.close()

my_database.close()
