import os
import sys
import re

allLinks = {}

def makeLink(fname, title):
	ss = title.lower().replace("#","").lstrip(' ').rstrip(' ').replace(' ','-').replace('=','').replace('&','').replace('`','').replace("'","").replace(":","").replace('"',"").replace('`', "").replace('.', "").replace('(',"").replace(')',"").replace("[","").replace("]","").replace("?","").replace("/","").replace("*","").strip("\n").rstrip("-")
	#ss.replace("--","-")
	#ss.replace(".", "")
	return fname+"#" + ss

def getLinksFromFile(curpath, fname):
	if not fname.endswith("md"):
		return
	fname = curpath + "/" + fname
	allLinks[fname] = "true" # each md file itself is a link
	allLinks[fname[:-3]] = "true" # each md file itself is a link, without .md is also ok
	print("processing file to get link"  + fname)
	with open(fname) as fread:
		cb = False # code block
		for x in fread:
			if x.lstrip(" ").startswith("```"):
				if cb == False:
					cb = True
				else:
					cb = False

			if cb: # code block
				continue

			if x[0] == '#':
				print(x)
				lk = makeLink(fname, x)
				#print(lk + " is a link")
				if "port-requirements-for" in lk:
					print(lk)
				if lk in allLinks:
					print("duplicated link")
				else:
					allLinks[lk] = "true"

class GetLinks:
	def __init__(self, idx, txt, curpath, fname):
		self._idx = idx
		self._txt = txt
		self._curpath = curpath
		self._fname = fname
		self._links = []
		self._keys = []
		self._original = []

	def getFlags(self):
		return [-1, -1, -1, -1]

	def resetFlags(self):
		return self.getFlags()

	def Process(self):
		if self._txt:
			flags = self.getFlags()
			for idx, x in enumerate(self._txt):
				if x == "[":
					if flags[0] < 0:
						flags[0] = idx
					else:
						flags = self.resetFlags() # invalid format
				elif x == "]":
					if flags[0] > -1: 
						flags[1] = idx # a valid []
				elif x == "(":
					if flags[1] > -1 and idx == flags[1] + 1: # [](
						flags[2] = idx
					else:
						flags = self.resetFlags()
				elif x == ")":
					if flags[2] > -1:
						flags[3] = idx # []()
						substr = self._txt[flags[0]:flags[3]+1]
						self._original.append(substr)
						#print(substr) # get on
						self._keys.append(self._txt[flags[0]+1:flags[1]])
						self._links.append(self._txt[flags[2]+1:flags[3]])
						flags = self.resetFlags()
					else:
						flags = self.resetFlags()

		#if len(self.links) > 0:
		#	print(self.links)

	def specialProcessOnLinks(self, lk):
			anchor = lk.find("/#")
			if anchor != -1:
				if not lk[:anchor].endswith(".md"):
					lk = lk[:anchor]+".md"+lk[anchor+1:]
			else:
				anchor = lk.find("#")
				if anchor != -1:
					if not lk[:anchor].endswith(".md"):
						lk = lk[:anchor]+".md"+lk[anchor:]
			return lk

	def stripLastPath(self, curpath):
		anchor = curpath.rfind("/")
		if anchor == -1:
			print("error processing path " + curpath +" cannot find parent path")
			return curpath
		return curpath[:anchor]


	def generatePath(self, curpath, curlink):
		return curpath + "/" + curlink


	def backTracePath(self, curpath, curlink):
		# back trace current
		parentpath = self.stripLastPath(curpath)
		nextlink = curlink[3:]
		# back trace parent
		if nextlink.startswith("../"):
			return self.backTracePath(parentpath, nextlink)
		else:
			return self.generatePath(parentpath, nextlink)

	def checkLinks(self):
		for idx, x in enumerate(self._links):
			if x == " ":
				continue
			if x.startswith("../"): #
				#lk = self.stripLastPath(self._curpath) + "/" + x[3:] # strip "../"
				lk = self.backTracePath(self._curpath, x)
				lk = self.specialProcessOnLinks(lk)
				if not lk in allLinks:
					print(self._fname + ":" + str(self._idx) + " link " + self._original[idx] + " generated link " + lk + " is not found")				
			elif x.startswith("./"): #relative link
				lk = self._curpath + "/" + x[2:] # strip "./"
				lk = self.specialProcessOnLinks(lk)
				if not lk in allLinks:
					print(self._fname + ":" + str(self._idx) + " link " + self._original[idx] + " generated link " + lk + " is not found")
			elif x.startswith("#"): #link to self
				lk = self._fname + x
				if not lk in allLinks:
					print(self._fname + ":" + str(self._idx) + " link " + self._original[idx] + " is not found")

def checkLinksOnFile(curpath, fname):
	if not fname.endswith("md"):
		return
	fname = curpath + "/" + fname
	print("processing file to validate link"  + fname)
	with open(fname) as fread:
		for idx, x in enumerate(fread):
			#link = re.match(r'.*\[.*\]\(.*#.*\).*',x)
			if x == "":
				continue
			gl = GetLinks(idx, x, curpath, fname)
			gl.Process()
			gl.checkLinks()


class WalkMd:
	def __init__(self, processFunc):
		self._process = processFunc

	def walkFiles(self, curpath, files):
		for f in files:
			self._process(curpath, f)

	def walkSubFolders(self, curPath, subFolders):
		for sf in subFolders:
			rpath = curPath + sf
			for newroot, newSubFolders, files in os.walk(rpath):
				self.walkFiles(newroot, files)
				self.walkSubFolders(newroot, newSubFolders)

	def walkRoot(self, rootdir):
		for root, subFolders, files in os.walk(rootdir):
			self.walkFiles(root, files)
			self.walkSubFolders(root, subFolders)


if __name__ == "__main__":
	if len(sys.argv) == 1:
		print("set a path to check")
	else:
		wmGetLinks = WalkMd(getLinksFromFile)
		wmGetLinks.walkRoot(sys.argv[1])

		print("link count " + str(len(allLinks)))

		wmCheckLinks = WalkMd(checkLinksOnFile)
		wmCheckLinks.walkRoot(sys.argv[1])	
