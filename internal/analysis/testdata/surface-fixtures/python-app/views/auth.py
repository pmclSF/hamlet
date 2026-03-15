from flask import request, jsonify

@app.post('/api/login')
def login_handler():
    email = request.json.get('email')
    return jsonify({'token': 'abc123'})

@app.get('/api/users')
def list_users():
    return jsonify([])

def validate_token(token):
    return token is not None
