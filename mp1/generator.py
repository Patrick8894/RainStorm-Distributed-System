import sys
import random

random.seed(int(sys.argv[1]))
print("Seed:", int(sys.argv[1]))

# Function to generate fixed data
def fixed_data():
    return """21.5.186.114 - - [17/Aug/2022:20:18:39 -0500] "POST /app/main/posts HTTP/1.0" 200 4993 "http://www.walker-freeman.com/home/" "Opera/8.43.(Windows NT 6.0; it-IT) Presto/2.9.185 Version/12.00"
58.122.255.161 - - [17/Aug/2022:20:19:12 -0500] "PUT /app/main/posts HTTP/1.0" 200 4998 "http://www.lyons.info/explore/home.htm" "Mozilla/5.0 (X11; Linux i686) AppleWebKit/5332 (KHTML, like Gecko) Chrome/15.0.860.0 Safari/5332"
189.150.240.255 - - [17/Aug/2022:20:22:27 -0500] "PUT /wp-admin HTTP/1.0" 404 4979 "http://www.jackson.net/tags/search/login.htm" "Mozilla/5.0 (Macintosh; U; PPC Mac OS X 10_8_6 rv:6.0; it-IT) AppleWebKit/534.27.6 (KHTML, like Gecko) Version/5.0 Safari/534.27.6"
246.99.38.108 - - [17/Aug/2022:20:24:45 -0500] "GET /wp-admin HTTP/1.0" 200 4944 "http://www.dean.net/" "Mozilla/5.0 (Macintosh; PPC Mac OS X 10_7_1; rv:1.9.3.20) Gecko/2021-11-18 13:56:56 Firefox/3.8"
140.92.134.207 - - [17/Aug/2022:20:26:55 -0500] "GET /wp-admin HTTP/1.0" 200 4972 "http://www.johnson.com/register.htm" "Mozilla/5.0 (Windows 95; it-IT; rv:1.9.1.20) Gecko/2012-12-18 09:46:43 Firefox/3.8"
198.112.45.0 - - [17/Aug/2022:20:30:46 -0500] "DELETE /apps/cart.jsp?appID=4464 HTTP/1.0" 200 4981 "http://www.robertson.com/search/tags/post.asp" "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_7_2; rv:1.9.6.20) Gecko/2018-03-24 17:49:58 Firefox/3.6.19"
25.236.12.45 - - [17/Aug/2022:20:31:54 -0500] "GET /app/main/posts HTTP/1.0" 200 5011 "http://anderson.com/search/category/" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_8_9; rv:1.9.5.20) Gecko/2017-02-07 06:07:22 Firefox/3.6.16"
182.26.25.156 - - [17/Aug/2022:20:35:27 -0500] "GET /wp-content HTTP/1.0" 200 5052 "http://www.walls-lawson.net/" "Mozilla/5.0 (Windows NT 6.1; it-IT; rv:1.9.2.20) Gecko/2020-06-28 07:31:03 Firefox/7.0"
58.216.243.158 - - [17/Aug/2022:20:38:53 -0500] "DELETE /wp-content HTTP/1.0" 200 5033 "http://brown.net/list/home/" "Mozilla/5.0 (X11; Linux i686) AppleWebKit/5322 (KHTML, like Gecko) Chrome/13.0.852.0 Safari/5322"
148.255.213.123 - - [17/Aug/2022:20:40:05 -0500] "GET /list HTTP/1.0" 200 4914 "http://thompson-sanchez.biz/author.htm" "Mozilla/5.0 (Windows NT 6.2; sl-SI; rv:1.9.2.20) Gecko/2012-01-26 01:56:48 Firefox/3.6.10"
        """

def generate_ip():
    return f'{random.randint(0, 255)}.{random.randint(0, 255)}.{random.randint(0, 255)}.{random.randint(0, 255)}'

def generate_timestamp():
    return f'{random.randint(1, 31)}/Aug/2022:{random.randint(0, 23)}:{random.randint(0, 59)}:{random.randint(0, 59)} -0500'

def generate_http_method():
    methods = ["GET", "POST", "PUT", "DELETE"]
    return random.choice(methods)

def generate_url():
    urls = ["/app/main/posts", "/wp-admin", "/apps/cart.jsp?appID=4464", "/app/main/posts", "/wp-content", "/list"]
    return random.choice(urls)

def generate_status_code():
    status_codes = [200, 404, 500]
    return random.choice(status_codes)

def generate_user_agent():
    user_agents = [
        "Mozilla/5.0",
        "Opera/8.43.(Windows NT 6.0; it-IT) Presto/2.9.185 Version/12.00", 
        "Mozilla/5.0 (X11; Linux i686) AppleWebKit/5332 (KHTML, like Gecko) Chrome/15.0.860.0 Safari/5332",
        "Mozilla/5.0 (Macintosh; U; PPC Mac OS X 10_8_6 rv:6.0; it-IT) AppleWebKit/534.27.6 (KHTML, like Gecko) Version/5.0 Safari/534.27.6",
        "Mozilla/5.0 (Macintosh; PPC Mac OS X 10_7_1; rv:"
        "Mozilla/5.0 (Windows 95; it-IT; rv:"
        "Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_7_2; rv:"
    ]
    return random.choice(user_agents)



# Function to generate log entry
def generate_log_entry():
    ip = generate_ip()
    timestamp = generate_timestamp()
    method = generate_http_method()
    url = generate_url()
    status_code = generate_status_code()
    size = random.randint(4900, 5100)
    referrer = generate_url()
    user_agent = generate_user_agent()
    
    return f'{ip} - - [{timestamp}] "{method} {url} HTTP/1.0" {status_code} {size} "{referrer}" "{user_agent}"'

# Generate log file with 10 entries
for i in range(10):
    with open(f"data/test_vm{i+1}.log", "w") as log_file:
        log_file.write(fixed_data() + "\n")
        for j in range(10):
            log_file.write(generate_log_entry() + "\n")

print("Log file generated successfully.")