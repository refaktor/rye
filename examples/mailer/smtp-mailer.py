from email.mime.text import MIMEText
from email.mime.application import MIMEApplication
from email.mime.multipart import MIMEMultipart
from smtplib import SMTP_SSL
from email import Charset
import MySQLdb
import time

db = MySQLdb.connect(host="localhost", user="demo", passwd="password", db=database)

smtp = SMTP("smtp.server.com")
smtp.ehlo()
smtp.login(smtp_user, smtp_pwd)

def run_mailer(database, sender, smtp_user, smtp_pwd):
    r = None
    cursor = db.cursor ()
    cursor.execute ("select * from mailer where status < 3 order by id limit 1")
    if cursor.rowcount > 0:
	r = cursor.fetchone ()
	cursor.close ()
        try:
	    send_message(
	        subj=r[6], from_=r[4], reply_to=r[5], to=r[3], content=r[7], \
	        file=r[8], filename=r[9], file2=r[10], filename2=r[11],
	        smtp_user=smtp_user, smtp_pwd=smtp_pwd )
	    cursor2 = db.cursor ()
	    cursor2.execute ("update mailer set status = 100 where id = " + str(r[0]) + ";")
	    cursor2.close()
	except:
	    cursor3 = db.cursor ()
	    cursor3.execute ("update mailer set status = status + 1 where id = " + str(r[0]) + ";")
	    cursor3.close()
	    raise
	if r is None: break

def send_message(subj, from_, to, content, file):
    
    msg = MIMEMultipart()
    msg['Subject'] = subj
    msg['From'] = from_
    msg['To'] = to

    part = MIMEText(content, 'plain', 'UTF-8')
    msg.attach(part)

    if file:
	part = MIMEApplication(open(file,"rb").read())
	part.add_header('Content-Disposition', 'attachment', filename=file)
	msg.attach(part)
        
    return msg

def send_message(subj, from_, to, content, file):

    msg = construct_message(subj, from_, to, content, file)
    smtp.sendmail(msg['From'], msg['To'], msg.as_string())

for range(1,5):
    run_mailer()
    time.sleep(2)
	