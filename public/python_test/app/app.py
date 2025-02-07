from flask import Flask

app = Flask(__name__)

@app.route("/")
def index():
    return "Hello, this domain is fully handled by PythonPlugin!"

@app.route("/test")
def test():
    return "Another route under Python's control."
