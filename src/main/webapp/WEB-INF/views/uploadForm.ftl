<html>
<head>
    <meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
    <title>file upload</title>
</head>
<body>
<p>${message}</p>
<form action="upload.html" method="post" enctype="multipart/form-data">
<p>
    <label for="text">server path</label> 
    <input type="text" name="path"> <br/>
    <label for="file">file</label>
    <input type="file" name="file">
</p>
<p>
    <input type="submit" value="submit"/>
</p>
</form>
</body>
</html>