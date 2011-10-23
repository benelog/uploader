package net.benelog.uploader;

import java.io.File;
import java.io.IOException;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.multipart.MultipartFile;

/**
 * Handles requests for the application home page.
 */
@Controller
public class UploadController {

	private static final Logger logger = LoggerFactory.getLogger(UploadController.class);
	static final String FORM_VIEW_NAME = "uploadForm";

	
	@RequestMapping("/form")
	public String form(Map<String, Object> out) {
		putMessage(out, "ready!");
		return FORM_VIEW_NAME;

	}
	
	@RequestMapping("/upload")
	public String upload(@RequestParam MultipartFile file, String path, Map<String, Object> out) throws IOException {

		String message = null;
		if (file.isEmpty()) {
			logger.info("empty file");
			message = "empty file";
		} else {
			String fileName = file.getOriginalFilename();
			File dest = new File(path + fileName);
			file.transferTo(dest);
			message = createMessage(file, dest);
			logger.info(message);
			
		}
		putMessage(out, message);
		return FORM_VIEW_NAME;
	}


	private String createMessage(MultipartFile file, File dest) {
		String writedPath = dest.getAbsolutePath();
		long size = file.getSize();
		String message = String.format("%s (%dbytes) uploaded!", writedPath,size);
		return message;
	}
	private void putMessage(Map<String, Object> out, String message) {
		out.put("message", message);
	}
}
