from bs4 import BeautifulSoup
from pymongo import MongoClient
import requests
import time

MAIN_URL = "https://codingbat.com"
RUN_URL = "https://codingbat.com/run"

client = MongoClient("localhost", 27017)

class Problem:

    def __init__(self, prob_id, prob_set, prob_name, prob_desc, prob_tests, prob_code, next_prob, last_prob):
        """
        Problem Model
        - problem_id: ID of problem (ex. p187868)
        - problem_set: The problem set name/identifier that the problem belongs to (ex. Warmup-1)
        - problem_name: The name of the problem (ex. sleepIn)
        - problem_description: The description of the problem
        - problem_tests: The tests that are run to check the problem code
        - problem_start_code:  The starting code for the problem
        - next_problem: The problem ID of the next problem in the set (None if last problem)
        - prev_problem: The problem ID of the previous problem in the set (None if first problem)
        """

        self.problem_id = prob_id 
        self.problem_set = prob_set 
        self.problem_name = prob_name 
        self.problem_description = prob_desc 
        self.problem_tests = prob_tests
        self.problem_start_code = prob_code
        self.next_problem = next_prob
        self.prev_problem = last_prob

    def to_dict(self):
        return {
            "problem_id": self.problem_id,
            "problem_set": self.problem_set,
            "problem_name": self.problem_name,
            "problem_description": self.problem_description,
            "problem_tests": self.problem_tests,
            "problem_start_code": self.problem_start_code,
            "next_problem": self.next_problem,
            "prev_problem": self.prev_problem, 
        }

    def __repr__(self):
        return str(self.to_dict())

    def __str__(self):
        return str(self.to_dict())

def get_problem_sets():
    time.sleep(1)
    problem_sets = []

    a = requests.get(MAIN_URL)
    index_html = a.content
    soup = BeautifulSoup(index_html, "html.parser")

    divs = soup.find_all("div", class_="summ")
    for d in divs:
        desc = list(d.strings)[-1].strip()
        problem_set = d.find('a')
        problem_sets.append({
            "problem_set_name": problem_set.string, 
            "problem_set_url": problem_set.get('href'),
            "problem_set_description": desc})

    return problem_sets

def get_problems(problem_set_url):
    time.sleep(1)
    problems = []

    problems_list_url = "%s%s" % (MAIN_URL, problem_set_url)
    a = requests.get(problems_list_url)
    problem_set_html = a.content
    soup = BeautifulSoup(problem_set_html, "html.parser")

    problem_div = soup.find("div", class_="indent")
    table = problem_div.find("table")
    problem_links = table.find_all("a")
    for pl in problem_links:
        problems.append({
            "problem_name": pl.string, 
            "problem_id": pl.get('href')})

    return problems

def get_problem_info(prob_url, prob_name, prob_set):
    time.sleep(1)

    problem_url = "%s%s" % (MAIN_URL, prob_url)
    a = requests.get(problem_url)
    problem_html = a.content
    soup = BeautifulSoup(problem_html, "html.parser")

    # Get problem description
    desc_list = []
    prob_desc = soup.find("p", class_="max2")
    examples = prob_desc.find_all_next(string=True)
    for ex in examples:
        if ex == "Go":
            break
        desc_list.append(ex.string)
    description = "\n".join(desc_list)

    # Get next and prev problems
    next_prob, prev_prob = "", ""

    next_ = soup.find("a", string="next")
    if next_ is not None:
        next_prob = next_.get("href")

    prev_ = soup.find("a", string="prev")
    if prev_ is not None:
        prev_prob = prev_.get("href")

    # Get problem tests
    form = soup.find("div", id="ace_div")
    code_snippet = form.string
    run_code = make_test_code(code_snippet)
    tests = get_problem_tests(prob_url, run_code)

    problem = Problem(prob_url, prob_set, prob_name, description, tests, code_snippet, next_prob, prev_prob)

    return problem

def get_problem_tests(prob_id, code):
    time.sleep(1)

    tests = []

    data = {
        "id": prob_url_to_id(prob_id),
        "code": code,
        "cuname": "",
        "owner": "",
        "adate": "20190516-185438z",
    }

    a = requests.post(RUN_URL, data=data)
    tests_html = a.content
    soup = BeautifulSoup(tests_html, "html.parser")
    table_rows = soup.find_all("tr")
    for tr in table_rows[1:]:
        table_data = tr.find("td")
        if "other tests" not in table_data.string:
            tests.append(table_data.string)

    return tests

def prob_url_to_id(prob_url):
    return prob_url[6:]

def make_test_code(code_snippet):
    return_code = "return"
    code_split = code_snippet.split(" ")
    return_type = code_split[1]
    if return_type == "boolean":
        return_code += " true;"
    elif return_type == "String":
        return_code += " \"\";"
    elif return_type == "int":
        return_code += " 0;"
    else:
        return_code += " null;"

    code_split[-2] = return_code
    return " ".join(code_split)

def run():
    problem_sets = get_problem_sets()
    problems = {}

    db = client["goding-bat"]
    ps = db.problem_sets 
    pm = db.problem_map
    probs = db.problem_info

    for problem_set in problem_sets:
        # Save problem set to DB
        ps.update_one({"problem_set_name": problem_set["problem_set_name"]}, 
                      {"$set": problem_set}, upsert=True)
        print(problem_set["problem_set_name"])

        # TODO: save using prob id rather than prob url?
        problems = get_problems(problem_set["problem_set_url"])
        # Save problem map to DB
        pm.update_one({"name": problem_set["problem_set_name"]}, {"$set": 
            {"name": problem_set["problem_set_name"], 
             "description": problem_set["problem_set_description"], 
             "problems": problems}}, upsert=True)

        for p in problems:
            print("\t", p["problem_name"], p["problem_id"])
            try:
                prob_info = get_problem_info(p["problem_id"], p["problem_name"], 
                                             problem_set["problem_set_name"])
                # Save problem to DB
                probs.update_one({"problem_id": p["problem_id"]}, 
                                 {"$set": prob_info.to_dict()}, upsert=True)
            except Exception as e:
                print("error:", p["problem_name"], p["problem_id"])
                print(e)


if __name__ == "__main__":
    run()
