// https://stackoverflow.com/questions/31458753/how-to-send-emails-form-database-with-java

import java.sql.Connection;
	import java.sql.DriverManager;
	import java.sql.PreparedStatement;
	import java.sql.ResultSet;
	import java.sql.SQLException;
	import java.util.ArrayList;
	import java.util.List;
	import java.util.Properties;

	import javax.mail.Message;
	import javax.mail.MessagingException;
	import javax.mail.PasswordAuthentication;
	import javax.mail.Session;
	import javax.mail.Transport;
	import javax.mail.internet.InternetAddress;
	import javax.mail.internet.MimeMessage;

	class Employee {
	    private String emailTo;
	    private String emailSubject;
	    private String emailBody;
	    private String emailAttachments;

	    public Employee() {
	        // TODO Auto-generated constructor stub
	    }

	    public Employee(String emailTo, String emailSubject, String emailBody,
	            String emailAttachments) {
	        super();
	        this.emailTo = emailTo;
	        this.emailSubject = emailSubject;
	        this.emailBody = emailBody;
	        this.emailAttachments = emailAttachments;
	    }

	    public String getEmailTo() {
	        return emailTo;
	    }

	    public void setEmailTo(String emailTo) {
	        this.emailTo = emailTo;
	    }

	    public String getEmailSubject() {
	        return emailSubject;
	    }

	    public void setEmailSubject(String emailSubject) {
	        this.emailSubject = emailSubject;
	    }

	    public String getEmailBody() {
	        return emailBody;
	    }

	    public void setEmailBody(String emailBody) {
	        this.emailBody = emailBody;
	    }

	    public String getEmailAttachments() {
	        return emailAttachments;
	    }

	    public void setEmailAttachments(String emailAttachments) {
	        this.emailAttachments = emailAttachments;
	    }

	}

	class EmployeeDao {
	    private Connection con;

	    private static final String GET_EMPLOYEES = "Select * From Employees";

	    private void connect() throws InstantiationException,
	            IllegalAccessException, ClassNotFoundException, SQLException {
	        Class.forName("com.microsoft.sqlserver.jdbc.SQLServerDriver")
	                .newInstance();
	        con = DriverManager
	                .getConnection("jdbc:sqlserver://100.00.000.000\\SQLEXPRESS:3316;databaseName=Employee");
	    }

	    public List<Employee> getEmployees() throws Exception {
	        connect();
	        PreparedStatement ps = con.prepareStatement(GET_EMPLOYEES);
	        ResultSet rs = ps.executeQuery();
	        List<Employee> result = new ArrayList<Employee>();
	        while (rs.next()) {
	            result.add(new Employee(rs.getString("emailTo"), rs
	                    .getString("emailSubject"), rs.getString("emailBody"), rs
	                    .getString("emailAttachments")));
	        }
	        disconnect();
	        return result;
	    }

	    private void disconnect() throws SQLException {
	        if (con != null) {
	            con.close();
	        }
	    }
	}

	class EmailSender {
	    private Session session;

	    private void init() {
	        Properties props = new Properties();
	        props.put("mail.smtp.auth", "true");
	        props.put("mail.smtp.starttls.enable", "true");
	        props.put("mail.smtp.host", "100.00.000.000");
	        props.put("mail.smtp.port", "20");

	        session = Session.getInstance(props, new javax.mail.Authenticator() {
	            protected PasswordAuthentication getPasswordAuthentication() {
	                return new PasswordAuthentication("work@gmail.com", "1234");
	            }
	        });
	    }

	    public void sendEmail(Employee e) throws MessagingException {
	         init();
	         Message message = new MimeMessage(session);
	         message.setFrom(new InternetAddress("work@gmail.com"));
	         message.setRecipients(Message.RecipientType.TO,
	             InternetAddress.parse(e.getEmailTo()));
	         message.setSubject(e.getEmailSubject());
	         message.setText(e.getEmailBody());
	         Transport.send(message);
	    }
	    public void sendEmail(List<Employee> employees) throws MessagingException{
	        for (Employee employee : employees) {
'	            sendEmail(employee);
	        }
	    }
	}

	public class Main {
	    public static void main(String[] args) throws Exception {
	        EmployeeDao dao=new EmployeeDao();
	        List<Employee> list=dao.getEmployees();
	        EmailSender sender=new EmailSender();
	        sender.sendEmail(list);
	    }
	}