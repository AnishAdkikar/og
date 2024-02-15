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

    # This returns in string format seperated by space
    output=''
    result=[]
    for row in cur:
        for i in row:    
            output+=str(i)+" "
        result.append(output)
        output=""
    return jsonify(result)

if __name__ == '__main__':
    app.run(debug=True)