import time as t 
import os
from win10toast import ToastNotifier #delete if you don't have win10toast installed!!!

notify = ToastNotifier() #delete if you don't have win10toast installed!!!

hours = 0
minutes = 0
seconds = 0
x = 0

def check():
    if minutes > 60:
        print("Wrong amount!!!")
        t.sleep(1)
    elif seconds > 60:
        print("Wrong amount!!!")
        t.sleep(1)
    else:
        return 1
while x != 1:
    os.system('cls')
    print("Remember to switch off quick_edit!!!")
    print("In what time do you want your PC to be switched off?")
    hours = int(input("Hours: "))
    minutes = int(input("Minutes: "))
    x = check()
    if x == 1: 
        seconds = int(input("Seconds: "))
        x = check()
    
while seconds != -1:
    if hours > 0:
        if minutes > 0:
            if seconds > 0:
                seconds -=1
                t.sleep(1)
            else: 
                minutes -=1
                seconds = 59
        elif minutes == 0:
            if seconds > 0:
                seconds -=1
                t.sleep(1)
            else:
                hours -=1
                minutes = 60
    elif hours == 0 and minutes > 0:
        if seconds > 0:
                seconds -=1
                t.sleep(1)
        else: 
            minutes -=1
            seconds = 59
    elif hours == 0 and minutes == 0 and seconds >= 0:
        seconds -=1
        t.sleep(1)
    
    if seconds >= 0: 
        os.system('cls')
        print("Time left: "+str(hours)+":"+str(minutes)+":"+str(seconds))
        print("\nTo stop press CTRL+C")

    if hours == 0 and minutes == 5 and seconds == 0:
        notify.show_toast(title="5 minutes to shutdown left", msg="Timer", threaded=True, duration=240) #delete if you don't have win10toast installed!!!
    
    if hours == 0 and minutes == 0 and seconds == 30:
        notify.show_toast(title="30 seconds to shutdown left!", msg="Timer", threaded=True, duration=30) #delete if you don't have win10toast installed!!!

    if hours == 0 and minutes == 0 and seconds == 0:
        os.system('shutdown /s /t 1')