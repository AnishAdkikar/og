from flask import Flask, jsonify
from flask_mysqldb import MySQL

app = Flask(__name__)

# Configure MySQL connection
app.config['MYSQL_USER'] = 'root'
app.config['MYSQL_PASSWORD'] = '1qwerty9'
app.config['MYSQL_DB'] = 'nodal_officers'


mysql = MySQL(app)

@app.route('/')
def index():
    cur = mysql.connection.cursor()
    cur.execute("SELECT * FROM list_nodal_officer ")
    
    # This returns in dictionary format
    # output = [dict((cur.description[i][0], value) for i,value in enumerate(row)) for row in cur.fetchall()]
    # return jsonify({'results': output}) 
   
    # output=[row for row in cur.fetchall() ]
    # return jsonify(output)

    
    output = ""
    result = {} # returns each row with its index
    final_result={} # returns group of rows with 6 elements of result
    index = 1

    for row in cur:
        sub_result = []
        for i in row:
            output = output + str(i) + " "
        sub_result.append(output)
        output=''
        
        result[index] = sub_result
        index += 1
    
    count=0
    for i in range(1,(len(result)//6)+1):
        sub_result=[]
        while count<i*6:
            sub_result.append(result[count+1])
            count+=1
            
        final_result[i] = sub_result
        
    return jsonify(final_result)

if __name__ == '__main__':
    app.run(debug=True)